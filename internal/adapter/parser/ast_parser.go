package parser

import (
	"fmt"
	"go/ast"
	"go/token"

	"goapianalyzer/internal/core/domain/entity"
	"goapianalyzer/pkg/errors"
	"goapianalyzer/pkg/utils"
)

type ASTParser struct {
	fileSet *token.FileSet
}

func NewASTParser(fileSet *token.FileSet) *ASTParser {
	return &ASTParser{
		fileSet: fileSet,
	}
}

func (p *ASTParser) ParseFileDetails(fileInfo *entity.FileInfo) error {
	if fileInfo == nil || fileInfo.AST == nil {
		return errors.NewValidationError("file info or AST is nil")
	}

	// Parse all declarations in the file
	for _, decl := range fileInfo.AST.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			funcInfo := p.parseFunctionDecl(d, fileInfo)
			fileInfo.Functions = append(fileInfo.Functions, funcInfo)
		case *ast.GenDecl:
			p.parseGenDecl(d, fileInfo)
		}
	}

	return nil
}

func (p *ASTParser) parseFunctionDecl(funcDecl *ast.FuncDecl, fileInfo *entity.FileInfo) *entity.FunctionInfo {
	funcInfo := &entity.FunctionInfo{
		Name:       funcDecl.Name.Name,
		Parameters: make([]*entity.Parameter, 0),
		Returns:    make([]*entity.Return, 0),
		Body:       p.getNodeText(funcDecl, fileInfo.Content),
		CallsTo:    make([]*entity.FunctionCall, 0),
		UsedTypes:  make([]string, 0),
	}

	// Get receiver if it's a method
	if funcDecl.Recv != nil && len(funcDecl.Recv.List) > 0 {
		recv := funcDecl.Recv.List[0]
		funcInfo.Receiver = p.exprToString(recv.Type)
		funcInfo.IsMethod = true
	}

	// Parse parameters
	if funcDecl.Type.Params != nil {
		for _, field := range funcDecl.Type.Params.List {
			paramType := p.exprToString(field.Type)
			if len(field.Names) == 0 {
				// Anonymous parameter
				funcInfo.Parameters = append(funcInfo.Parameters, &entity.Parameter{
					Name: "",
					Type: paramType,
				})
			} else {
				for _, name := range field.Names {
					funcInfo.Parameters = append(funcInfo.Parameters, &entity.Parameter{
						Name: name.Name,
						Type: paramType,
					})
				}
			}
		}
	}

	// Parse return types
	if funcDecl.Type.Results != nil {
		for _, field := range funcDecl.Type.Results.List {
			returnType := p.exprToString(field.Type)
			if len(field.Names) == 0 {
				// Anonymous return
				funcInfo.Returns = append(funcInfo.Returns, &entity.Return{
					Name: "",
					Type: returnType,
				})
			} else {
				for _, name := range field.Names {
					funcInfo.Returns = append(funcInfo.Returns, &entity.Return{
						Name: name.Name,
						Type: returnType,
					})
				}
			}
		}
	}

	// Parse function body for function calls and used types
	if funcDecl.Body != nil {
		p.parseBlockStmt(funcDecl.Body, funcInfo, fileInfo)
	}

	return funcInfo
}

func (p *ASTParser) parseGenDecl(genDecl *ast.GenDecl, fileInfo *entity.FileInfo) {
	switch genDecl.Tok {
	case token.TYPE:
		p.parseTypeDecl(genDecl, fileInfo)
	case token.VAR:
		p.parseVarDecl(genDecl, fileInfo)
	case token.CONST:
		p.parseConstDecl(genDecl, fileInfo)
	}
}

func (p *ASTParser) parseTypeDecl(genDecl *ast.GenDecl, fileInfo *entity.FileInfo) {
	for _, spec := range genDecl.Specs {
		if typeSpec, ok := spec.(*ast.TypeSpec); ok {
			typeName := typeSpec.Name.Name

			switch t := typeSpec.Type.(type) {
			case *ast.StructType:
				structInfo := p.parseStructType(typeName, t, fileInfo)
				fileInfo.Structs = append(fileInfo.Structs, structInfo)
				fileInfo.Types = append(fileInfo.Types, &entity.TypeInfo{
					Name: typeName,
					Type: "struct",
					Body: p.getNodeText(typeSpec, fileInfo.Content),
				})
			case *ast.InterfaceType:
				interfaceInfo := p.parseInterfaceType(typeName, t, fileInfo)
				fileInfo.Interfaces = append(fileInfo.Interfaces, interfaceInfo)
				fileInfo.Types = append(fileInfo.Types, &entity.TypeInfo{
					Name: typeName,
					Type: "interface",
					Body: p.getNodeText(typeSpec, fileInfo.Content),
				})
			default:
				fileInfo.Types = append(fileInfo.Types, &entity.TypeInfo{
					Name: typeName,
					Type: p.exprToString(t),
					Body: p.getNodeText(typeSpec, fileInfo.Content),
				})
			}
		}
	}
}

func (p *ASTParser) parseStructType(name string, structType *ast.StructType, fileInfo *entity.FileInfo) *entity.StructInfo {
	structInfo := &entity.StructInfo{
		Name:   name,
		Fields: make([]*entity.StructField, 0),
		Body:   p.getNodeText(structType, fileInfo.Content),
	}

	if structType.Fields != nil {
		for _, field := range structType.Fields.List {
			fieldType := p.exprToString(field.Type)
			tag := ""
			if field.Tag != nil {
				tag = field.Tag.Value
			}

			if len(field.Names) == 0 {
				// Embedded field
				structInfo.Fields = append(structInfo.Fields, &entity.StructField{
					Name:     "",
					Type:     fieldType,
					Tag:      tag,
					Embedded: true,
				})
			} else {
				for _, name := range field.Names {
					structInfo.Fields = append(structInfo.Fields, &entity.StructField{
						Name:     name.Name,
						Type:     fieldType,
						Tag:      tag,
						Embedded: false,
					})
				}
			}
		}
	}

	return structInfo
}

func (p *ASTParser) parseInterfaceType(name string, interfaceType *ast.InterfaceType, fileInfo *entity.FileInfo) *entity.InterfaceInfo {
	interfaceInfo := &entity.InterfaceInfo{
		Name:    name,
		Methods: make([]*entity.InterfaceMethod, 0),
		Body:    p.getNodeText(interfaceType, fileInfo.Content),
	}

	if interfaceType.Methods != nil {
		for _, method := range interfaceType.Methods.List {
			if len(method.Names) > 0 {
				methodName := method.Names[0].Name
				if funcType, ok := method.Type.(*ast.FuncType); ok {
					interfaceMethod := &entity.InterfaceMethod{
						Name:       methodName,
						Parameters: make([]*entity.Parameter, 0),
						Returns:    make([]*entity.Return, 0),
					}

					// Parse parameters
					if funcType.Params != nil {
						for _, param := range funcType.Params.List {
							paramType := p.exprToString(param.Type)
							if len(param.Names) == 0 {
								interfaceMethod.Parameters = append(interfaceMethod.Parameters, &entity.Parameter{
									Name: "",
									Type: paramType,
								})
							} else {
								for _, paramName := range param.Names {
									interfaceMethod.Parameters = append(interfaceMethod.Parameters, &entity.Parameter{
										Name: paramName.Name,
										Type: paramType,
									})
								}
							}
						}
					}

					// Parse returns
					if funcType.Results != nil {
						for _, result := range funcType.Results.List {
							returnType := p.exprToString(result.Type)
							if len(result.Names) == 0 {
								interfaceMethod.Returns = append(interfaceMethod.Returns, &entity.Return{
									Name: "",
									Type: returnType,
								})
							} else {
								for _, returnName := range result.Names {
									interfaceMethod.Returns = append(interfaceMethod.Returns, &entity.Return{
										Name: returnName.Name,
										Type: returnType,
									})
								}
							}
						}
					}

					interfaceInfo.Methods = append(interfaceInfo.Methods, interfaceMethod)
				}
			}
		}
	}

	return interfaceInfo
}

func (p *ASTParser) parseVarDecl(genDecl *ast.GenDecl, fileInfo *entity.FileInfo) {
	for _, spec := range genDecl.Specs {
		if valueSpec, ok := spec.(*ast.ValueSpec); ok {
			varType := ""
			if valueSpec.Type != nil {
				varType = p.exprToString(valueSpec.Type)
			}

			for i, name := range valueSpec.Names {
				varInfo := &entity.VariableInfo{
					Name: name.Name,
					Type: varType,
					Body: p.getNodeText(valueSpec, fileInfo.Content),
				}

				if i < len(valueSpec.Values) {
					varInfo.Value = p.exprToString(valueSpec.Values[i])
				}

				fileInfo.Variables = append(fileInfo.Variables, varInfo)
			}
		}
	}
}

func (p *ASTParser) parseConstDecl(genDecl *ast.GenDecl, fileInfo *entity.FileInfo) {
	for _, spec := range genDecl.Specs {
		if valueSpec, ok := spec.(*ast.ValueSpec); ok {
			constType := ""
			if valueSpec.Type != nil {
				constType = p.exprToString(valueSpec.Type)
			}

			for i, name := range valueSpec.Names {
				constInfo := &entity.ConstantInfo{
					Name: name.Name,
					Type: constType,
					Body: p.getNodeText(valueSpec, fileInfo.Content),
				}

				if i < len(valueSpec.Values) {
					constInfo.Value = p.exprToString(valueSpec.Values[i])
				}

				fileInfo.Constants = append(fileInfo.Constants, constInfo)
			}
		}
	}
}

func (p *ASTParser) parseBlockStmt(block *ast.BlockStmt, funcInfo *entity.FunctionInfo, fileInfo *entity.FileInfo) {
	for _, stmt := range block.List {
		p.parseStmt(stmt, funcInfo, fileInfo)
	}
}

func (p *ASTParser) parseStmt(stmt ast.Stmt, funcInfo *entity.FunctionInfo, fileInfo *entity.FileInfo) {
	switch s := stmt.(type) {
	case *ast.ExprStmt:
		p.parseExpr(s.X, funcInfo, fileInfo)
	case *ast.AssignStmt:
		for _, expr := range s.Rhs {
			p.parseExpr(expr, funcInfo, fileInfo)
		}
	case *ast.IfStmt:
		if s.Cond != nil {
			p.parseExpr(s.Cond, funcInfo, fileInfo)
		}
		if s.Body != nil {
			p.parseBlockStmt(s.Body, funcInfo, fileInfo)
		}
		if s.Else != nil {
			p.parseStmt(s.Else, funcInfo, fileInfo)
		}
	case *ast.ForStmt:
		if s.Cond != nil {
			p.parseExpr(s.Cond, funcInfo, fileInfo)
		}
		if s.Body != nil {
			p.parseBlockStmt(s.Body, funcInfo, fileInfo)
		}
	case *ast.ReturnStmt:
		for _, result := range s.Results {
			p.parseExpr(result, funcInfo, fileInfo)
		}
	case *ast.BlockStmt:
		p.parseBlockStmt(s, funcInfo, fileInfo)
	}
}

func (p *ASTParser) parseExpr(expr ast.Expr, funcInfo *entity.FunctionInfo, fileInfo *entity.FileInfo) {
	switch e := expr.(type) {
	case *ast.CallExpr:
		p.parseCallExpr(e, funcInfo, fileInfo)
	case *ast.CompositeLit:
		if e.Type != nil {
			typeName := p.exprToString(e.Type)
			if !utils.Contains(funcInfo.UsedTypes, typeName) {
				funcInfo.UsedTypes = append(funcInfo.UsedTypes, typeName)
			}
		}
		for _, elt := range e.Elts {
			p.parseExpr(elt, funcInfo, fileInfo)
		}
	case *ast.SelectorExpr:
		p.parseExpr(e.X, funcInfo, fileInfo)
	case *ast.BinaryExpr:
		p.parseExpr(e.X, funcInfo, fileInfo)
		p.parseExpr(e.Y, funcInfo, fileInfo)
	case *ast.UnaryExpr:
		p.parseExpr(e.X, funcInfo, fileInfo)
	}
}

func (p *ASTParser) parseCallExpr(callExpr *ast.CallExpr, funcInfo *entity.FunctionInfo, fileInfo *entity.FileInfo) {
	funcCall := &entity.FunctionCall{
		Name:      p.exprToString(callExpr.Fun),
		Arguments: make([]string, 0),
		Position:  p.getPosition(callExpr.Pos()),
	}

	for _, arg := range callExpr.Args {
		funcCall.Arguments = append(funcCall.Arguments, p.exprToString(arg))
	}

	funcInfo.CallsTo = append(funcInfo.CallsTo, funcCall)

	// Also parse arguments for nested expressions
	for _, arg := range callExpr.Args {
		p.parseExpr(arg, funcInfo, fileInfo)
	}
}

func (p *ASTParser) exprToString(expr ast.Expr) string {
	if expr == nil {
		return ""
	}

	switch e := expr.(type) {
	case *ast.Ident:
		return e.Name
	case *ast.SelectorExpr:
		return p.exprToString(e.X) + "." + e.Sel.Name
	case *ast.StarExpr:
		return "*" + p.exprToString(e.X)
	case *ast.ArrayType:
		return "[]" + p.exprToString(e.Elt)
	case *ast.MapType:
		return "map[" + p.exprToString(e.Key) + "]" + p.exprToString(e.Value)
	case *ast.ChanType:
		return "chan " + p.exprToString(e.Value)
	case *ast.InterfaceType:
		return "interface{}"
	case *ast.StructType:
		return "struct{...}"
	case *ast.FuncType:
		return "func(...)"
	case *ast.BasicLit:
		return e.Value
	case *ast.CompositeLit:
		return p.exprToString(e.Type) + "{...}"
	case *ast.CallExpr:
		return p.exprToString(e.Fun) + "(...)"
	default:
		return fmt.Sprintf("%T", e)
	}
}

func (p *ASTParser) getNodeText(node ast.Node, content string) string {
	if p.fileSet == nil {
		return ""
	}

	pos := p.fileSet.Position(node.Pos())
	end := p.fileSet.Position(node.End())

	if pos.Offset >= len(content) || end.Offset > len(content) {
		return ""
	}

	return content[pos.Offset:end.Offset]
}

func (p *ASTParser) getPosition(pos token.Pos) *entity.Position {
	if p.fileSet == nil {
		return nil
	}

	position := p.fileSet.Position(pos)
	return &entity.Position{
		Line:   position.Line,
		Column: position.Column,
		Offset: position.Offset,
	}
}

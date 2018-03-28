package semantic_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/ifql/ast"
	"github.com/influxdata/ifql/semantic"
	"github.com/influxdata/ifql/semantic/semantictest"
)

func TestNew(t *testing.T) {
	testCases := []struct {
		name    string
		program *ast.Program
		want    *semantic.Program
		wantErr bool
	}{
		{
			name:    "empty",
			program: &ast.Program{},
			want: &semantic.Program{
				Body: []semantic.Statement{},
			},
		},
		{
			name: "var declaration",
			program: &ast.Program{
				Body: []ast.Statement{
					&ast.VariableDeclaration{
						Declarations: []*ast.VariableDeclarator{{
							ID:   &ast.Identifier{Name: "a"},
							Init: &ast.BooleanLiteral{Value: true},
						}},
					},
					&ast.ExpressionStatement{
						Expression: &ast.Identifier{Name: "a"},
					},
				},
			},
			want: &semantic.Program{
				Body: []semantic.Statement{
					&semantic.NativeVariableDeclaration{
						Identifier: &semantic.Identifier{Name: "a"},
						Init:       &semantic.BooleanLiteral{Value: true},
					},
					&semantic.ExpressionStatement{
						Expression: &semantic.IdentifierExpression{Name: "a"},
					},
				},
			},
		},
		{
			name: "function",
			program: &ast.Program{
				Body: []ast.Statement{
					&ast.VariableDeclaration{
						Declarations: []*ast.VariableDeclarator{{
							ID: &ast.Identifier{Name: "f"},
							Init: &ast.ArrowFunctionExpression{
								Params: []*ast.Property{
									{Key: &ast.Identifier{Name: "a"}},
									{Key: &ast.Identifier{Name: "b"}},
								},
								Body: &ast.BinaryExpression{
									Operator: ast.AdditionOperator,
									Left:     &ast.Identifier{Name: "a"},
									Right:    &ast.Identifier{Name: "b"},
								},
							},
						}},
					},
					&ast.ExpressionStatement{
						Expression: &ast.CallExpression{
							Callee: &ast.Identifier{Name: "f"},
							Arguments: []ast.Expression{&ast.ObjectExpression{
								Properties: []*ast.Property{
									{Key: &ast.Identifier{Name: "a"}, Value: &ast.IntegerLiteral{Value: 2}},
									{Key: &ast.Identifier{Name: "b"}, Value: &ast.IntegerLiteral{Value: 3}},
								},
							}},
						},
					},
				},
			},
			want: &semantic.Program{
				Body: []semantic.Statement{
					&semantic.NativeVariableDeclaration{
						Identifier: &semantic.Identifier{Name: "f"},
						Init: &semantic.FunctionExpression{
							Params: []*semantic.FunctionParam{
								{Key: &semantic.Identifier{Name: "a"}},
								{Key: &semantic.Identifier{Name: "b"}},
							},
							Body: &semantic.BinaryExpression{
								Operator: ast.AdditionOperator,
								Left: &semantic.IdentifierExpression{
									Name: "a",
								},
								Right: &semantic.IdentifierExpression{
									Name: "b",
								},
							},
						},
					},
					&semantic.ExpressionStatement{
						Expression: &semantic.CallExpression{
							Callee: &semantic.IdentifierExpression{
								Name: "f",
							},
							Arguments: &semantic.ObjectExpression{
								Properties: []*semantic.Property{
									{Key: &semantic.Identifier{Name: "a"}, Value: &semantic.IntegerLiteral{Value: 2}},
									{Key: &semantic.Identifier{Name: "b"}, Value: &semantic.IntegerLiteral{Value: 3}},
								},
							},
						},
					},
				},
			},
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got, err := semantic.New(tc.program, nil)
			if !tc.wantErr && err != nil {
				t.Fatal(err)
			} else if tc.wantErr && err == nil {
				t.Fatal("expected error")
			}

			if !cmp.Equal(tc.want, got, semantictest.CmpOptions...) {
				t.Errorf("unexpected semantic program: -want/+got:\n%s", cmp.Diff(tc.want, got, semantictest.CmpOptions...))
			}
		})
	}
}

func TestExpression_Kind(t *testing.T) {
	testCases := []struct {
		name string
		expr semantic.Expression
		want semantic.Kind
	}{
		{
			name: "string",
			expr: &semantic.StringLiteral{},
			want: semantic.String,
		},
		{
			name: "int",
			expr: &semantic.IntegerLiteral{},
			want: semantic.Int,
		},
		{
			name: "uint",
			expr: &semantic.UnsignedIntegerLiteral{},
			want: semantic.UInt,
		},
		{
			name: "float",
			expr: &semantic.FloatLiteral{},
			want: semantic.Float,
		},
		{
			name: "bool",
			expr: &semantic.BooleanLiteral{},
			want: semantic.Bool,
		},
		{
			name: "time",
			expr: &semantic.DateTimeLiteral{},
			want: semantic.Time,
		},
		{
			name: "duration",
			expr: &semantic.DurationLiteral{},
			want: semantic.Duration,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := tc.expr.Type().Kind()

			if !cmp.Equal(tc.want, got) {
				t.Errorf("unexpected expression type: -want/+got:\n%s", cmp.Diff(tc.want, got))
			}
		})
	}
}

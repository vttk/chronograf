package server

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/chronograf"
	"github.com/influxdata/chronograf/mocks"
	"github.com/influxdata/chronograf/organizations"
	"github.com/influxdata/chronograf/roles"
)

func TestStore_SourcesGet(t *testing.T) {
	type fields struct {
		SourcesStore chronograf.SourcesStore
	}
	type args struct {
		superAdmin   bool
		organization string
		role         string
		id           int
	}
	type wants struct {
		source chronograf.Source
		err    bool
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		wants  wants
	}{
		{
			name: "Get viewer source as viewer",
			fields: fields{
				SourcesStore: &mocks.SourcesStore{
					GetF: func(ctx context.Context, id int) (chronograf.Source, error) {
						return chronograf.Source{
							ID:           1,
							Name:         "my sweet name",
							Organization: "0",
							Role:         "viewer",
						}, nil
					},
				},
			},
			args: args{
				organization: "0",
				role:         "viewer",
			},
			wants: wants{
				source: chronograf.Source{
					ID:           1,
					Name:         "my sweet name",
					Organization: "0",
					Role:         "viewer",
				},
			},
		},
		{
			name: "Get viewer source as editor",
			fields: fields{
				SourcesStore: &mocks.SourcesStore{
					GetF: func(ctx context.Context, id int) (chronograf.Source, error) {
						return chronograf.Source{
							ID:           1,
							Name:         "my sweet name",
							Organization: "0",
							Role:         "viewer",
						}, nil
					},
				},
			},
			args: args{
				organization: "0",
				role:         "editor",
			},
			wants: wants{
				source: chronograf.Source{
					ID:           1,
					Name:         "my sweet name",
					Organization: "0",
					Role:         "viewer",
				},
			},
		},
		{
			name: "Get viewer source as admin",
			fields: fields{
				SourcesStore: &mocks.SourcesStore{
					GetF: func(ctx context.Context, id int) (chronograf.Source, error) {
						return chronograf.Source{
							ID:           1,
							Name:         "my sweet name",
							Organization: "0",
							Role:         "viewer",
						}, nil
					},
				},
			},
			args: args{
				organization: "0",
				role:         "admin",
			},
			wants: wants{
				source: chronograf.Source{
					ID:           1,
					Name:         "my sweet name",
					Organization: "0",
					Role:         "viewer",
				},
			},
		},
		{
			name: "Get admin source as viewer",
			fields: fields{
				SourcesStore: &mocks.SourcesStore{
					GetF: func(ctx context.Context, id int) (chronograf.Source, error) {
						return chronograf.Source{
							ID:           1,
							Name:         "my sweet name",
							Organization: "0",
							Role:         "admin",
						}, nil
					},
				},
			},
			args: args{
				organization: "0",
				role:         "viewer",
			},
			wants: wants{
				err: true,
			},
		},
		{
			name: "Get editor source as viewer",
			fields: fields{
				SourcesStore: &mocks.SourcesStore{
					GetF: func(ctx context.Context, id int) (chronograf.Source, error) {
						return chronograf.Source{
							ID:           1,
							Name:         "my sweet name",
							Organization: "0",
							Role:         "editor",
						}, nil
					},
				},
			},
			args: args{
				organization: "0",
				role:         "viewer",
			},
			wants: wants{
				err: true,
			},
		},
		{
			name: "Get editor source as editor",
			fields: fields{
				SourcesStore: &mocks.SourcesStore{
					GetF: func(ctx context.Context, id int) (chronograf.Source, error) {
						return chronograf.Source{
							ID:           1,
							Name:         "my sweet name",
							Organization: "0",
							Role:         "editor",
						}, nil
					},
				},
			},
			args: args{
				organization: "0",
				role:         "editor",
			},
			wants: wants{
				source: chronograf.Source{
					ID:           1,
					Name:         "my sweet name",
					Organization: "0",
					Role:         "editor",
				},
			},
		},
		{
			name: "Get editor source as admin",
			fields: fields{
				SourcesStore: &mocks.SourcesStore{
					GetF: func(ctx context.Context, id int) (chronograf.Source, error) {
						return chronograf.Source{
							ID:           1,
							Name:         "my sweet name",
							Organization: "0",
							Role:         "editor",
						}, nil
					},
				},
			},
			args: args{
				organization: "0",
				role:         "admin",
			},
			wants: wants{
				source: chronograf.Source{
					ID:           1,
					Name:         "my sweet name",
					Organization: "0",
					Role:         "editor",
				},
			},
		},
		{
			name: "Get editor source as viewer",
			fields: fields{
				SourcesStore: &mocks.SourcesStore{
					GetF: func(ctx context.Context, id int) (chronograf.Source, error) {
						return chronograf.Source{
							ID:           1,
							Name:         "my sweet name",
							Organization: "0",
							Role:         "editor",
						}, nil
					},
				},
			},
			args: args{
				organization: "0",
				role:         "viewer",
			},
			wants: wants{
				err: true,
			},
		},
		{
			name: "Get admin source as admin",
			fields: fields{
				SourcesStore: &mocks.SourcesStore{
					GetF: func(ctx context.Context, id int) (chronograf.Source, error) {
						return chronograf.Source{
							ID:           1,
							Name:         "my sweet name",
							Organization: "0",
							Role:         "admin",
						}, nil
					},
				},
			},
			args: args{
				organization: "0",
				role:         "admin",
			},
			wants: wants{
				source: chronograf.Source{
					ID:           1,
					Name:         "my sweet name",
					Organization: "0",
					Role:         "admin",
				},
			},
		},
		{
			name: "Get source as super admin",
			fields: fields{
				SourcesStore: &mocks.SourcesStore{
					GetF: func(ctx context.Context, id int) (chronograf.Source, error) {
						return chronograf.Source{
							ID:           1,
							Name:         "my sweet name",
							Organization: "0",
							Role:         "viewer",
						}, nil
					},
				},
			},
			args: args{
				organization: "0",
				role:         "viewer",
				superAdmin:   true,
			},
			wants: wants{
				source: chronograf.Source{
					ID:           1,
					Name:         "my sweet name",
					Organization: "0",
					Role:         "viewer",
				},
			},
		},
		{
			name: "Get source as super admin - no organization or role",
			fields: fields{
				SourcesStore: &mocks.SourcesStore{
					GetF: func(ctx context.Context, id int) (chronograf.Source, error) {
						return chronograf.Source{
							ID:           1,
							Name:         "my sweet name",
							Organization: "0",
							Role:         "viewer",
						}, nil
					},
				},
			},
			args: args{
				superAdmin: true,
			},
			wants: wants{
				err: true,
			},
		},
		{
			name: "Get source as super admin - no role",
			fields: fields{
				SourcesStore: &mocks.SourcesStore{
					GetF: func(ctx context.Context, id int) (chronograf.Source, error) {
						return chronograf.Source{
							ID:           1,
							Name:         "my sweet name",
							Organization: "0",
							Role:         "viewer",
						}, nil
					},
				},
			},
			args: args{
				superAdmin:   true,
				organization: "0",
			},
			wants: wants{
				err: true,
			},
		},
		{
			name: "Get source as super admin - no organization",
			fields: fields{
				SourcesStore: &mocks.SourcesStore{
					GetF: func(ctx context.Context, id int) (chronograf.Source, error) {
						return chronograf.Source{
							ID:           1,
							Name:         "my sweet name",
							Organization: "0",
							Role:         "viewer",
						}, nil
					},
				},
			},
			args: args{
				superAdmin: true,
				role:       "viewer",
			},
			wants: wants{
				err: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &Store{
				SourcesStore: tt.fields.SourcesStore,
			}

			ctx := context.Background()

			if tt.args.superAdmin {
				ctx = context.WithValue(ctx, SuperAdminKey, true)
			}

			if tt.args.organization != "" {
				ctx = context.WithValue(ctx, organizations.ContextKey, tt.args.organization)
			}

			if tt.args.role != "" {
				ctx = context.WithValue(ctx, roles.ContextKey, tt.args.role)
			}

			source, err := store.Sources(ctx).Get(ctx, tt.args.id)
			if (err != nil) != tt.wants.err {
				t.Errorf("%q. Store.Sources().Get() error = %v, wantErr %v", tt.name, err, tt.wants.err)
				return
			}
			if diff := cmp.Diff(source, tt.wants.source); diff != "" {
				t.Errorf("%q. Store.Sources().Get():\n-got/+want\ndiff %s", tt.name, diff)
			}
		})
	}
}

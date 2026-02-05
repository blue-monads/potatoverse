package ownorm

type FindManyOptions struct {
	TableName string
	Cond      map[string]any
	Select    []string
	OrderBy   []string
	Limit     int64
	Offset    int64
}

type OwnORM interface {
	FindMany(options *FindManyOptions) (any, error)
	FindOne(options any) (any, error)
	CreateMany(options any) (any, error)
	CreateOne(options any) (any, error)
	UpdateMany(options any) (any, error)
	UpdateOne(options any) (any, error)
	DeleteMany(options any) (any, error)
	DeleteOne(options any) (any, error)
}

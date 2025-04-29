package contracts

type GetNameQuery struct {
	ID int
}

type GetNameQueryResponse struct {
	ID       int
	UserName string
}

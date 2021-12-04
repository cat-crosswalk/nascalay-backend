package handler

type handler struct{}

func NewHandler() ServerInterface {
	return &handler{}
}

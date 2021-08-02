package services

type IService interface {
	Run()
}

type ServiceBase struct{}

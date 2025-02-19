package usecase

type Service struct {
	UserUseCase  *UserUseCase
	TourUseCase  *TourismUseCase
	AdminUseCase *AdminUseCase
}

func NewService(user *UserUseCase, tour *TourismUseCase, admin *AdminUseCase) *Service {
	return &Service{
		UserUseCase:  user,
		TourUseCase:  tour,
		AdminUseCase: admin,
	}
}

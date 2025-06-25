package test_suites

import (
	"context"
	container "duolingo/libraries/service_container"
	"duolingo/libraries/service_container/test/fake"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
)

type ServiceContainerTestSuite struct {
	suite.Suite
}

func (s *ServiceContainerTestSuite) SetupSuite() {
	container.Init(context.Background())
}

func (s *ServiceContainerTestSuite) Test_Bind_Alias_Transient() {
	container.BindAlias("dog_as_animal", func(ctx context.Context) any {
		return &fake.Dog{Id: uuid.NewString()}
	})

	firstDog, firstResolve := container.ResolveAlias[fake.Animal]("dog_as_animal")
	secondDog, secondResolved := container.ResolveAlias[fake.Animal]("dog_as_animal")

	s.Assert().True(firstResolve)
	s.Assert().NotNil(firstDog)

	s.Assert().True(secondResolved)
	s.Assert().NotNil(secondDog)

	s.Assert().NotSame(firstDog, secondDog)
	s.Assert().NotEqual(firstDog.MakeSound(), secondDog.MakeSound())
}

func (s *ServiceContainerTestSuite) Test_Bind_Alias_Singleton() {
	container.BindSingletonAlias("dog_as_animal", func(ctx context.Context) any {
		return &fake.Dog{Id: uuid.NewString()}
	})
	firstDog, firstResolve := container.ResolveAlias[fake.Animal]("dog_as_animal")
	secondDog, secondResolved := container.ResolveAlias[fake.Animal]("dog_as_animal")

	s.Assert().True(firstResolve)
	s.Assert().NotNil(firstDog)

	s.Assert().True(secondResolved)
	s.Assert().NotNil(secondDog)

	s.Assert().Same(firstDog, secondDog)
	s.Assert().Equal(firstDog.MakeSound(), secondDog.MakeSound())
}

func (s *ServiceContainerTestSuite) Test_Bind_Transient() {
	container.Bind[fake.Animal](func(ctx context.Context) any {
		return &fake.Dog{Id: uuid.NewString()}
	})

	firstDog, firstResolve := container.Resolve[fake.Animal]()
	secondDog, secondResolved := container.Resolve[fake.Animal]()

	s.Assert().True(firstResolve)
	s.Assert().NotNil(firstDog)

	s.Assert().True(secondResolved)
	s.Assert().NotNil(secondDog)

	s.Assert().NotSame(firstDog, secondDog)
	s.Assert().NotEqual(firstDog.MakeSound(), secondDog.MakeSound())
}

func (s *ServiceContainerTestSuite) Test_Bind_Singleton() {
	container.BindSingleton[fake.Animal](func(ctx context.Context) any {
		return &fake.Dog{Id: uuid.NewString()}
	})
	firstDog, firstResolve := container.Resolve[fake.Animal]()
	secondDog, secondResolved := container.Resolve[fake.Animal]()

	s.Assert().True(firstResolve)
	s.Assert().NotNil(firstDog)

	s.Assert().True(secondResolved)
	s.Assert().NotNil(secondDog)

	s.Assert().Same(firstDog, secondDog)
	s.Assert().Equal(firstDog.MakeSound(), secondDog.MakeSound())
}

func (s *ServiceContainerTestSuite) Test_Bind_Concrete() {
	dog := &fake.Dog{Id: uuid.NewString()}
	container.BindSingleton[*fake.Dog](func(ctx context.Context) any {
		return dog
	})
	resolved, ok := container.Resolve[*fake.Dog]()
	s.Assert().True(ok)
	s.Assert().Equal(resolved.MakeSound(), dog.MakeSound())
}

func (s *ServiceContainerTestSuite) Test_Bind_Mismatch() {
	// mismatch pointer
	container.Bind[fake.Dog](func(ctx context.Context) any {
		return &fake.Dog{Id: uuid.NewString()}
	})
	_, firstResolve := container.Resolve[fake.Dog]()

	// mismatch type
	container.Bind[*fake.Dog](func(ctx context.Context) any {
		return &fake.Cat{Id: "cat_" + uuid.NewString()}
	})
	_, secondResolve := container.Resolve[*fake.Dog]()

	// nil
	container.Bind[fake.Animal](func(ctx context.Context) any {
		return nil
	})
	_, thirdResolve := container.Resolve[fake.Animal]()

	s.Assert().False(firstResolve)
	s.Assert().False(secondResolve)
	s.Assert().False(thirdResolve)
}

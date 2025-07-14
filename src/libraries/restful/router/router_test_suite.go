package router

import (
	"github.com/stretchr/testify/suite"
)

type RouterTestSuite struct {
	suite.Suite
	router *Router
}

func (s *RouterTestSuite) SetupTest() {
	s.router = NewRouter()
}

func (s *RouterTestSuite) Test_Router_InvalidRoutes() {
	err1 := s.router.Add("", "handler1")
	err2 := s.router.Add("/", "handler1")
	err3 := s.router.Add("/user/{id", "handler1")

	s.Assert().Equal(ErrAddEmptyRoute, err1)
	s.Assert().Equal(ErrAddEmptyRoute, err2)
	s.Assert().Equal(ErrArgumentNotEnclosed, err3)
}

func (s *RouterTestSuite) Test_Router_DuplicationErr() {
	err1 := s.router.Add("/user/{id}", "handler1")
	err2 := s.router.Add("/user/profile", "handler2")
	s.NoError(err1)
	s.Equal(err2, ErrRouteDuplication)
}

func (s *RouterTestSuite) Test_Router_ValidRoutes() {
	err1 := s.router.Add("/user", "dummy handler")
	err2 := s.router.Add("/user/{id}", "dummy handler")
	err3 := s.router.Add("/user/{id}/profile", "dummy handler")
	err4 := s.router.Add("/*", "dummy handler")
	s.NoError(err1)
	s.NoError(err2)
	s.NoError(err3)
	s.NoError(err4)
}

func (s *RouterTestSuite) Test_RouteResult_BasicCases() {
	s.Require().NoError(s.router.Add("/user/{id}", "H1"))
	s.Require().NoError(s.router.Add("/courses/{lang}/docs", "H2"))
	s.Require().NoError(s.router.Add("/*", "H3"))
	s.Require().NoError(s.router.Add("/user/{id}/profile", "H4"))

	m1, r1 := s.router.Match("/user/123")
	s.Assert().True(m1)
	s.Assert().Equal("H1", r1.GetHandler())
	s.Assert().Equal(map[string]string{"id": "123"}, r1.GetPathArgs())

	m2, r2 := s.router.Match("/courses/fr/docs")
	s.Assert().True(m2)
	s.Assert().Equal("H2", r2.GetHandler())
	s.Assert().Equal(map[string]string{"lang": "fr"}, r2.GetPathArgs())

	m3, r3 := s.router.Match("/*")
	s.Assert().True(m3)
	s.Assert().Equal("H3", r3.GetHandler())
	s.Assert().Equal(map[string]string{}, r3.GetPathArgs())

	m4, r4 := s.router.Match("/user/123/profile")
	s.Assert().True(m4)
	s.Assert().Equal("H4", r4.GetHandler())
	s.Assert().Equal(map[string]string{"id": "123"}, r4.GetPathArgs())
}

func (s *RouterTestSuite) Test_RouteResult_EdgeCases() {
	s.Require().NoError(s.router.Add("/user/{id}", "H1"))

	m1, r1 := s.router.Match("/product")
	s.Assert().False(m1)
	s.Assert().Nil(r1)

	m2, r2 := s.router.Match("/user/123/")
	s.Assert().True(m2)
	s.Assert().Equal("H1", r2.GetHandler())
	s.Assert().Equal(map[string]string{"id": "123"}, r2.GetPathArgs())

	m3, r3 := s.router.Match("/user")
	s.Assert().False(m3)
	s.Assert().Nil(r3)

	m4, r4 := s.router.Match("/user/123/posts")
	s.Assert().False(m4)
	s.Assert().Nil(r4)

	m5, r5 := s.router.Match("/user/a!b@c#d$e")
	s.Assert().True(m5)
	s.Assert().Equal("H1", r5.GetHandler())
	s.Assert().Equal(map[string]string{"id": "a!b@c#d$e"}, r5.GetPathArgs())
}

func (s *RouterTestSuite) Test_PathCleaning() {
	s.router.Add("//extra//slashes//", "CLEANED")
	matched, result := s.router.Match("//extra/slashes//")
	s.True(matched)
	s.Equal("extra/slashes", result.routeNode.pattern)
}

func (s *RouterTestSuite) Test_ArgumentExtraction() {
	s.router.Add("/{first}/{second}", "PAIR")
	matched, result := s.router.Match("/one/two")
	s.True(matched)
	s.Equal(
		map[string]string{"first": "one", "second": "two"},
		result.GetPathArgs(),
	)
}

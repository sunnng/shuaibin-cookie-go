package collect

import "app/internal/logger"

// Route navigates into/out of collect flows. Skeleton: no-ops.
type Route struct {
	page *Page
}

func NewRoute(page *Page) *Route {
	return &Route{page: page}
}

func (r *Route) Enter() bool {
	logger.Infof("[CollectRoute] Enter skeleton (no-op)")
	return true
}

func (r *Route) Leave() bool {
	logger.Infof("[CollectRoute] Leave skeleton (no-op)")
	return true
}

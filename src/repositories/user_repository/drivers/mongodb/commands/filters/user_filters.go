package filters

import (
	b "go.mongodb.org/mongo-driver/bson"
	"time"
)

type UserFilters struct {
	filters b.M
}

func NewUserFilters() *UserFilters {
	return &UserFilters{
		filters: b.M{},
	}
}

func (f *UserFilters) SetFilterIds(ids []string) {
	f.filters["user_id"] = b.M{"$in": ids}
}

func (f *UserFilters) SetFilterCampaign(campaign string) {
	f.filters["campaigns"] = campaign
}

func (f *UserFilters) SetFilterOnlyEmailVerified() {
	f.filters["$and"] = []b.M{
		{"email_verified_at": b.M{"$ne": nil}},
		{"email_verified_at": b.M{"$lte": time.Now().UTC()}},
	}
}

func (f *UserFilters) GetFilters() b.M {
	return f.filters
}

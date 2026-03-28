package eveonline

type Character struct {
	Id            int
	Name          string `json:"name"`
	CorporationId int64  `json:"corporation_id"`
}

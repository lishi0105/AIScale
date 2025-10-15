package types

// JSON Body: { "id": "<uuid>" }
type IDReq struct {
	ID string `json:"id" binding:"required,uuid4"`
}

package repository

type Review struct {
	UserID  string `bson:"userID"`
	MovieID string `bson:"movieID"`
	Text    string `bson:"text"`
}

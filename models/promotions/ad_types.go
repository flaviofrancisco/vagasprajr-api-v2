package promotions

type AdItem struct {
	ShortUrl    string `bson:"short_url" json:"short_url"`
	OriginalUrl string `bson:"original_url" json:"original_url"`
	QtyClicks   int    `bson:"qty_clicks" json:"qty_clicks"`
	Source      string `bson:"source" json:"source"`
}
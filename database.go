package seller

type SellerDetails struct {
	UserName string
	Password string
	Location string
	ItemName string
}

type ItemDetails struct {
	ItemName          string
	QuantityAvailable int
	Cost              float64
}

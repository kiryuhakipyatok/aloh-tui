package setups

type UsersSetup struct {
	VolumeCoefficient   float32 `json:"volume-coeficent"`
	Muted               bool    `json:"muted"`
	HardDenoise         bool    `json:"hardDenoise"`
	SoftDenoise         bool    `json:"softDenoise"`
	AmountOfConnections uint    `json:"amount-of-connections"`
}
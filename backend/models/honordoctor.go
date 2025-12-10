package models

type HonorDoctor struct {
	DoctorName              string
	CareproviderTxnDoctorId int64
	HonorFinal              float64
	TotalHonor              float64
	CountedMonth            int64
	CountedYear             int64
}

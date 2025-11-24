package models

type Patientbill struct {
	ID           int64
	VisitNo      string
	PatientName  string
	PatientType  string
	PatientClass string
	TxnCode      string
	TxnCategory  string
	TxnDesc      string
	TxnDoctor    string
	RegnDoctor   string
}

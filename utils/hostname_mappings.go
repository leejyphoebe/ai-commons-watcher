package utils

type Cluster string

const (
	Aspire2A  Cluster = "aspire2a"
	Aspire2AP Cluster = "aspire2ap"
	CCDSCPU   Cluster = "ccds_cpu"
	CCDSB200  Cluster = "ccds_b200"
	FourIR    Cluster = "4ir"
	Unknown   Cluster = "unknown"
)

func GetClusterFromHostname(hostname string) Cluster {
	// Extract the cluster name from the hostname
	switch {
	case hostname == "aspire2antu.nscc.sg":
		return Aspire2A
	case hostname == "aspire2pntu.nscc.sg":
		return Aspire2AP
	case hostname == "10.96.190.65":
		return CCDSCPU
	case hostname == "10.96.176.228":
		return FourIR
	case hostname == "10.96.191.22":
		return CCDSB200
	}
	return Unknown
}

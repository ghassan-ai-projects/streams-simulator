package domain

// Digests of the delivered stream for each perturbation alone and for all of them
// together over the 60-event golden stream (golden_test.go). They were recorded
// before the transform dispatch refactor (modularity round R6) and must not
// change without a reviewed perturbation change.
var goldenDigests = map[string]string{
	"clock_skew":      "883f1fdd04646989eb9d8ca3514c4ff35f79834e85bbad99bd6650eb756ea17d",
	"delay_tail":      "dc6862f9f6657108dd168cfd2d7a1978c8c20fdbf88cd57d6e0874b153b67c93",
	"drop":            "0b514e4ee8e85fec7c3465c3ed9ac70cab2b1b9c1d7e4d52c74e798a5a7ee991",
	"duplicate_burst": "68696a54234db1be260e97b44aca1ade01c9b25b7a3b3625af66e555e62df81c",
	"gross_backfill":  "49bda9ab99e292ebe8c5838f3ec31859f07893e7f6aad53645fef7954305b7be",
	"id_reuse":        "3cc0fffe22c03fc1a5826c8c64cc50ba030273afdf6422c971d00544bd82f855",
	"injection_probe": "bc70ace36a222402fd66fb116d95bdf2c38be44ee1eaf22985d1b8f0cf0ddae8",
	"malformed":       "9880e01ce336c0ae6a1352bc535bc9f686d0103b1aa7eb300c55517a36456d3f",
	"nan_inf":         "671122ce43cd68f0ff86e617fbb0c64391be011ec0094e630e2f270f36943d67",
	"non_monotonic":   "b2e3ffa266898f4c804d3953e1b3b6dfc5d91772ba63fe84dca69427249038cd",
	"out_of_enum":     "f23665603ea421380b33496289b3aee2a940094be2bc3a1936f313c098e39048",
	"out_of_range":    "93061ec7a93b4c852b2a8b86d997dc9da2d87f6b2cfcd5dd493de17343981bee",
	"oversize":        "0fdd958b5fffa41fb09b5dc5780c7c31d37c0cac50ac5d31daca06ae562dfdb2",
	"precision_edge":  "0f8a6431c9ef8eb6334e11c7103eb66c52fea8bf5b49bf0598ab0551a6e74177",
	"producer_flap":   "74234e98afe7498fb5daf1f36ac2d78acc339464f950703b8c019892f982b90b",
	"reorder":         "706303017af2a5eda18e60dba9d714f35729c3f4956dbf34bb07d4c7223d03c5",
	"storm":           "44c1fc3dc057d86e0cf7601f6bf54aca60692de10acfd56f9174026c581a26a0",
	"time_encoding":   "c7b03b9b53eb78f1762b4e297bce4b284e5589553d82c69a367b9e26665dc19a",
	"unit_mismatch":   "e8b51490864adb68a05dd999a0ded7cd0dc3a2780bc0920652946a80a1438c34",
}

const goldenAllDigest = "b1997b1ceec5de55524ae40de0bd30985e717874b38787a10566b88b4830e374"

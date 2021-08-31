package utils

import (
	"math/rand"
	"time"

	"github.com/ppc64le-cloud/pvsadm/pkg"
)

// Spec variables
var (
	planSlice   = []string{"smart", "standard", "vault", "cold"}
	regionSlice = []string{"us-east", "jp-tok", "us-south", "au-syd", "eu-de", "ca-tor"}
)

func randomInt(min, max int) int {
	return min + rand.Intn(max-min)
}

// Generate random string of given length
func GenerateRandomString(length int) string {
	rand.Seed(time.Now().UnixNano())
	bytes := make([]byte, length)
	for i := 0; i < length; i++ {
		bytes[i] = byte(randomInt(97, 122))
	}
	return string(bytes)
}

// Generate Specifications
func GenerateSpec(NoOfTargetsPerSource int) pkg.Spec {
	var spec pkg.Spec
	spec.Source = pkg.Source{
		Bucket:       "image-sync-" + GenerateRandomString(6),
		Cos:          "cos-image-sync-test-" + GenerateRandomString(6),
		Object:       "",
		StorageClass: planSlice[randomInt(0, len(planSlice))],
		Region:       regionSlice[randomInt(0, len(regionSlice))],
	}

	spec.Target = make([]pkg.TargetItem, 0)
	for tgt := 0; tgt < NoOfTargetsPerSource; tgt++ {
		spec.Target = append(spec.Target, pkg.TargetItem{
			Bucket:       "image-sync-" + GenerateRandomString(6),
			StorageClass: planSlice[randomInt(0, len(planSlice))],
			Region:       regionSlice[randomInt(0, len(regionSlice))],
		})
	}

	return spec
}

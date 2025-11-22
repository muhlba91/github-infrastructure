package google

import (
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/iam"
)

// WorkloadIdentityPool defines a Google workload identity pool.
type WorkloadIdentityPool struct {
	// WorkloadIdentityPool is the workload identity pool.
	WorkloadIdentityPool *iam.WorkloadIdentityPool
	// WorkloadIdentityProvider is the workload identity provider.
	WorkloadIdentityProvider *iam.WorkloadIdentityPoolProvider
}

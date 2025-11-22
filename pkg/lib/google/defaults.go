//nolint:gochecknoglobals // globals are allowed in this file
package google

// maxRepositoryLength defines the maximum length for the repository name used in GCP resource namings.
const maxRepositoryLength = 18

// postfixLength defines the length of the random postfix added to resource names for uniqueness.
const postfixLength = 8

// defaultPermissions defines the default set of GCP permissions assigned to service accounts.
var defaultPermissions = []string{
	"cloudkms.cryptoKeyVersions.useToDecrypt",
	"cloudkms.cryptoKeyVersions.useToEncrypt",
	"cloudkms.cryptoKeys.getIamPolicy",
	"cloudkms.cryptoKeys.setIamPolicy",
	"cloudkms.locations.get",
	"cloudkms.locations.list",
	"compute.regions.list",
	"iam.serviceAccountKeys.create",
	"iam.serviceAccountKeys.delete",
	"iam.serviceAccountKeys.disable",
	"iam.serviceAccountKeys.enable",
	"iam.serviceAccountKeys.get",
	"iam.serviceAccountKeys.list",
	"iam.serviceAccounts.create",
	"iam.serviceAccounts.delete",
	"iam.serviceAccounts.disable",
	"iam.serviceAccounts.enable",
	"iam.serviceAccounts.get",
	"iam.serviceAccounts.getIamPolicy",
	"iam.serviceAccounts.list",
	"iam.serviceAccounts.setIamPolicy",
	"iam.serviceAccounts.undelete",
	"iam.serviceAccounts.update",
	"resourcemanager.projects.get",
	"resourcemanager.projects.getIamPolicy",
	"resourcemanager.projects.setIamPolicy",
	"resourcemanager.projects.update",
	"storage.hmacKeys.create",
	"storage.hmacKeys.delete",
	"storage.hmacKeys.get",
	"storage.hmacKeys.list",
	"storage.hmacKeys.update",
	"storage.buckets.create",
	"storage.buckets.createTagBinding",
	"storage.buckets.delete",
	"storage.buckets.deleteTagBinding",
	"storage.buckets.get",
	"storage.buckets.getIamPolicy",
	"storage.buckets.getObjectInsights",
	"storage.buckets.list",
	"storage.buckets.listEffectiveTags",
	"storage.buckets.listTagBindings",
	"storage.buckets.setIamPolicy",
	"storage.buckets.update",
	"storage.multipartUploads.abort",
	"storage.multipartUploads.create",
	"storage.multipartUploads.list",
	"storage.multipartUploads.listParts",
	"storage.objects.create",
	"storage.objects.delete",
	"storage.objects.get",
	"storage.objects.getIamPolicy",
	"storage.objects.list",
	"storage.objects.setIamPolicy",
	"storage.objects.update",
}

// defaultServices defines the default set of GCP services to be enabled for projects.
var defaultServices = []string{
	"iam.googleapis.com",
	"iamcredentials.googleapis.com",
	"cloudresourcemanager.googleapis.com",
	"cloudkms.googleapis.com",
	"storage.googleapis.com",
	"storage-component.googleapis.com",
	"compute.googleapis.com",
}

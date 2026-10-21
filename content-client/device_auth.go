package contentclient

type DeviceAuthRequestStatus string

const (
	DeviceAuthRequestStatusPending  DeviceAuthRequestStatus = "pending"
	DeviceAuthRequestStatusApproved DeviceAuthRequestStatus = "approved"
	DeviceAuthRequestStatusDenied   DeviceAuthRequestStatus = "denied"
	DeviceAuthRequestStatusExpired  DeviceAuthRequestStatus = "expired"
	DeviceAuthRequestStatusConsumed DeviceAuthRequestStatus = "consumed"
)

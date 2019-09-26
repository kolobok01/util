package client

type Client interface {
	CreateCompatibleClient(onVersionSpecified, onVersionDetermined, onUsingDefaultVersion func(string)) (Client, error)
}

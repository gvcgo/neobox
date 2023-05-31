package errs

type UnSupportedProxySchemeError struct{}

func (that *UnSupportedProxySchemeError) Error() string {
	return "Unsupported Proxy Scheme"
}

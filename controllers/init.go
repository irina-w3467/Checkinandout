package controllers

// CCServer - root struct for the entire server
type CCServer struct {
	Validator CCValidator
	Config    Config
}

// InitServer - return the reference to a server instance
func InitServer() CCServer {
	return CCServer{}
}

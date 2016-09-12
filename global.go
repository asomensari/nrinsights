package nrinsights

// nilCheckClient is a special type of client that has nil protection build in, in case it isn't set up.
type nilCheckClient struct {
	client Client
}

// GlobalClient is a singleton global instance of the client
var globalClient nilCheckClient

//
var GlobalClient Client = &globalClient

// SetupGlobalClient sets up a the global singleton client using the provided details. If either the endpoint or token
// are empty, an error will be returned, and the global Client will silently no-op return on all methods.  If the
// global client has already been setup, an ErrDuplicateSetup will be returned.
func SetupGlobalClient(cfg Config) (err error) {
	if globalClient.client != nil {
		return ErrDuplicateSetup
	}
	globalClient.client, err = NewClient(cfg)
	return
}

// SetupGlobalClientFromFlags sets up a the global singleton client using the parsed flags. If either the endpoint or
// token are empty, an error will be returned, and the global Client will silently no-op return on all methods.  If the
// global client has already been setup, an ErrDuplicateSetup will be returned.
func SetupGlobalClientFromFlags() (err error) {
	if globalClient.client != nil {
		return ErrDuplicateSetup
	}
	globalClient.client, err = NewClientFromFlags()
	return
}

// Send uses the global instance of the nrinsights client
var Send = globalClient.Send

// Close uses the global instance of the nrinsights client
var Close = globalClient.Close

// Send calls the global client Send, but silently no-ops if the global client hasn't been setup
func (n *nilCheckClient) Send(message NREvent) {
	if n.client != nil {
		n.client.Send(message)
	}
}

// Close calls the global client Close, but silently no-ops if the global client hasn't been setup
func (n *nilCheckClient) Close() {
	if n.client != nil {
		n.client.Close()
	}
}

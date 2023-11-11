## Creating a New Plugin: Step-by-Step Instructions

* Define Plugin Interface
   - Example: `type Greeter interface { Greet() string }`

* Create RPC Implementation
   - Create an RPC structure (Example: `GreeterRPC`) that embeds an `*rpc.Client`
   - Implement the methods from your interface in this structure
   - For remote procedure calls:
     - Use `client.Call` method
     - Handle potential errors (consider returning errors or panicking)

* Define RPC Server
   - Create an RPC server structure (Example: `GreeterRPCServer`)
   - This server should embed the real implementation of the interface
   - Implement the same methods as your plugin interface in this server but make them RPC-compatible

* Implement Plugin Server/Client Methods
   - Create a structure (Example: `GreeterPlugin`) representing the plugin itself
   - This structure should embed your interface (Example: `Impl Greeter`)
   - Implement `Server` method:
     - Returns an instance of your RPC server structure
   - Implement `Client` method:
     - Returns an instance of your RPC implementation structure

* Omit Advanced Features (Initially)
    - Ignore `MuxBroker` unless implementing advanced features.

```go 
package shared

import (
	"net/rpc"
	"github.com/hashicorp/go-plugin"
)

type Greeter interface { Greet() string }

type GreeterRPC struct{ client *rpc.Client }

func (g *GreeterRPC) Greet() string {
	var resp string
	err := g.client.Call("Plugin.Greet", new(interface{}), &resp)
	if err != nil {
		panic(err)
	}

	return resp
}

type GreeterRPCServer struct {
	Impl Greeter
}

func (g *GreeterRPCServer) Greet(args interface{}, resp *string) error {
	*resp = g.Impl.Greet()
	return nil
}

type GreeterPlugin struct {
	Impl Greeter
}

func (p *GreeterPlugin) Server(broker *plugin.MuxBroker) (interface{}, error) {
	return &GreeterRPCServer{Impl: p.Impl}, nil
}

func (p *GreeterPlugin) Client(broker *plugin.MuxBroker, client *rpc.Client) (interface{}, error) {
	return &GreeterRPC{client: client}, nil
}
```

## Creating a New Plugin Implementation

* Implement Actual Plugin Logic
   - Create a real implementation of the interface (Example: `GreeterHello`)
   - Implement the methods defined in the interface
     - Add any desired logging or additional functionality

* Define Handshake Configuration
   - Create a `handshakeConfig` to set up a basic handshake between the plugin and the host
   - Set `ProtocolVersion`, `MagicCookieKey`, and `MagicCookieValue`

* Main Function Setup
   - Initialize a logger (Example: `hclog.New`)
   - Instantiate the real implementation of the interface and assign to a variable (Example: `greeter`)
   - Create a `pluginMap` to map the plugin interface to its real implementation

* Serve the Plugin
   - Call `plugin.Serve` function
   - Provide a `ServeConfig`:
     - Assign the `handshakeConfig`
     - Assign the `pluginMap` to the `Plugins` field

``` 
import (
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"github.com/pkg/errors"
	"net/rpc"
	"os"
)
```

```go 
package main

type Greeter struct {
	logger hclog.Logger
}

func (g *Greeter) Greet() string {
	g.logger.Info("Hello!")
	return "Hello!"
}

var handshakeConfig = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "GLAZED_PLUGIN",
	MagicCookieValue: "glazed",
}

func main() {
	logger := hclog.New(&hclog.LoggerOptions{
		Level:      hclog.Debug,
		Output:     os.Stderr,
		JSONFormat: true,
	})

	greeter := &Greeter{
		logger: logger,
	}

	var pluginMap = map[string]plugin.Plugin{
		// we could do a meld of the plugins here
		"greeter": greeter,
	}
	
	logger.Info("message from plugin")
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: handshakeConfig,
		Plugins:         pluginMap,
	})
}


```

## Loading and Using the Plugin

* Initialize Logger
   - Create an `hclog.Logger` for logging activities

* Launch the Plugin Process
   - Use `plugin.NewClient`:
     - Provide a `ClientConfig` with:
       - `HandshakeConfig` 
       - Plugin map (`Plugins`)
       - Command to start the plugin process (`Cmd`)
       - Logger 

* Connect to Plugin via RPC
   - Use `client.Client()` to establish an RPC connection
   - Handle any connection errors

* Request and Use the Plugin
   - Use `rpcClient.Dispense()` to request the desired plugin interface (Example: "greeter")
   - Cast the received plugin to its interface (Example: `raw.(shared.Greeter)`)
   - Invoke methods on the interface as if it was a local implementation (Example: `greeter.Greet()`)

* Clean Up
   - Use `client.Kill()` to close the client when done

```go
package main

import (
	"github.com/go-go-golems/go-go-labs/cmd/tests/plugin-test/shared"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"os"
	"os/exec"
)

func main() {
	logger := hclog.New(&hclog.LoggerOptions{
		Name:   "plugin",
		Output: os.Stderr,
		Level:  hclog.Debug,
	})

	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: plugin.HandshakeConfig{
			ProtocolVersion:  1,
			MagicCookieKey:   "GLAZED_PLUGIN",
			MagicCookieValue: "glazed",
		},
		Plugins: map[string]plugin.Plugin{
			"greeter":  &shared.GreeterPlugin{},
		},
		Cmd:    exec.Command("./plugin/greeter"),
		Logger: logger,
	})
	defer client.Kill()

	rpcClient, err := client.Client()
	if err != nil {
		logger.Error("Failed to create RPC client", "error", err)
		os.Exit(1)
	}

	raw, err := rpcClient.Dispense("greeter")
	if err != nil {
		logger.Error("Failed to dispense plugin", "error", err)
		os.Exit(1)
	}

	greeter := raw.(shared.Greeter)
	logger.Info("Calling plugin")
	greeting := greeter.Greet()
	logger.Info("Got response from plugin", "greeting", greeting)
}
```
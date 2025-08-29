# protoc-gen-route

`protoc-gen-route` is a protoc plugin that generates routing code from `.proto` files. It is designed to inspect service definitions within your protobuf files and automatically generate corresponding route handlers based on a specified template. This plugin creates Go code that provides structured routing with operation constants, extra data handling, server interfaces, and codec interfaces for seamless integration with various transport protocols.

## Features

- Generates operation constants for each service method
- Creates extra data mappings from proto options
- Provides server and codec interfaces for type-safe implementations
- Supports custom request and response models
- Generates handler functions with automatic request/response conversion
- Integrates with the sphere options framework
- Supports flexible template customization

## Installation

To install `protoc-gen-route`, use the following command:

```bash
go install github.com/go-sphere/protoc-gen-route@latest
```

## Prerequisites

You need to have the sphere options proto definitions in your project. Add the following dependency to your `buf.yaml`:

```yaml
deps:
  - buf.build/go-sphere/options
```

## Configuration Parameters

The behavior of `protoc-gen-route` can be customized with the following parameters:

- **`version`**: Print the current plugin version and exit. (Default: `false`)
- **`options_key`**: The key for the option extension in your proto file that contains routing information. (Default: `route`)
- **`file_suffix`**: The suffix for the generated files. (Default: `_route.pb.go`)
- **`template_file`**: Path to a custom Go template file. If not provided, the default internal template is used.
- **`request_model`**: (Required) The fully qualified Go type for the request model (e.g., `github.com/gin-gonic/gin.Context`).
- **`response_model`**: (Required) The fully qualified Go type for the response model.
- **`extra_data_model`**: The fully qualified Go type for an additional data model to be used in the template.
- **`extra_data_constructor`**: A function that constructs and returns a pointer to the `extra_data_model`. (Required if `extra_data_model` is set).

## Usage with Buf

To use `protoc-gen-route` with `buf`, you can configure it in your `buf.gen.yaml` file. Here is an example configuration:

```yaml
version: v2
managed:
  enabled: true
  disable:
    - file_option: go_package_prefix
      module: buf.build/go-sphere/options
plugins:
  - local: protoc-gen-go
    out: api
    opt: paths=source_relative
  - local: protoc-gen-route
    out: api
    opt:
      - paths=source_relative
      - options_key=bot
      - file_suffix=_bot.pb.go
      - request_model=github.com/go-sphere/sphere/social/telegram;Update
      - response_model=github.com/go-sphere/sphere/social/telegram;Message
      - extra_data_model=github.com/go-sphere/sphere/social/telegram;MethodExtraData
      - extra_data_constructor=github.com/go-sphere/sphere/social/telegram;NewMethodExtraData
```

## Proto Definition Example

Here's how to define services with routing options in your `.proto` files:

```protobuf
syntax = "proto3";

package bot.v1;

import "sphere/options/options.proto";

service MenuService {
  // UpdateCount handles count update operations
  // Supports both command and callback query triggers
  rpc UpdateCount(UpdateCountRequest) returns (UpdateCountResponse) {
    option (sphere.options.options) = {
      key: "bot"
      extra: [
        {
          key: "command"
          value: "start"
        },
        {
          key: "callback_query"
          value: "start"
        }
      ]
    };
  }
  
  // ProcessMenu handles menu navigation
  rpc ProcessMenu(ProcessMenuRequest) returns (ProcessMenuResponse) {
    option (sphere.options.options) = {
      key: "bot"
      extra: [
        {
          key: "callback_query"
          value: "menu_.*"
        }
      ]
    };
  }
}

message UpdateCountRequest {
  int64 value = 1;
  int64 offset = 2;
}

message UpdateCountResponse {
  int64 value = 1;
}

message ProcessMenuRequest {
  string menu_id = 1;
  string action = 2;
}

message ProcessMenuResponse {
  string result = 1;
}
```

## Generated Code

The plugin generates Go code with the following components for each service:

### Operation Constants

```go
const OperationBotMenuServiceUpdateCount = "/bot.v1.MenuService/UpdateCount"
const OperationBotMenuServiceProcessMenu = "/bot.v1.MenuService/ProcessMenu"
```

### Extra Data Variables

```go
var ExtraBotDataMenuServiceUpdateCount = telegram.NewMethodExtraData(map[string]string{
    "callback_query": "start",
    "command":        "start",
})

var ExtraBotDataMenuServiceProcessMenu = telegram.NewMethodExtraData(map[string]string{
    "callback_query": "menu_.*",
})
```

### Helper Functions

```go
func GetExtraBotDataByMenuServiceOperation(operation string) *telegram.MethodExtraData {
    switch operation {
    case OperationBotMenuServiceUpdateCount:
        return ExtraBotDataMenuServiceUpdateCount
    case OperationBotMenuServiceProcessMenu:
        return ExtraBotDataMenuServiceProcessMenu
    default:
        return nil
    }
}

func GetAllBotMenuServiceOperations() []string {
    return []string{
        OperationBotMenuServiceUpdateCount,
        OperationBotMenuServiceProcessMenu,
    }
}
```

### Server Interface

```go
type MenuServiceBotServer interface {
    // UpdateCount handles count update operations
    // Supports both command and callback query triggers
    UpdateCount(context.Context, *UpdateCountRequest) (*UpdateCountResponse, error)
    
    // ProcessMenu handles menu navigation
    ProcessMenu(context.Context, *ProcessMenuRequest) (*ProcessMenuResponse, error)
}
```

### Codec Interface

```go
type MenuServiceBotCodec interface {
    DecodeUpdateCountRequest(ctx context.Context, request *telegram.Update) (*UpdateCountRequest, error)
    EncodeUpdateCountResponse(ctx context.Context, response *UpdateCountResponse) (*telegram.Message, error)
    DecodeProcessMenuRequest(ctx context.Context, request *telegram.Update) (*ProcessMenuRequest, error)
    EncodeProcessMenuResponse(ctx context.Context, response *ProcessMenuResponse) (*telegram.Message, error)
}
```

### Handler Functions

```go
func _MenuService_UpdateCount0_Bot_Handler(srv MenuServiceBotServer, codec MenuServiceBotCodec, render func(ctx context.Context, request *telegram.Update, msg *telegram.Message) error) func(ctx context.Context, request *telegram.Update) error {
    return func(ctx context.Context, request *telegram.Update) error {
        req, err := codec.DecodeUpdateCountRequest(ctx, request)
        if err != nil {
            return err
        }
        resp, err := srv.UpdateCount(ctx, req)
        if err != nil {
            return err
        }
        msg, err := codec.EncodeUpdateCountResponse(ctx, resp)
        if err != nil {
            return err
        }
        return render(ctx, request, msg)
    }
}
```

### Registration Function

```go
func RegisterMenuServiceBotServer(srv MenuServiceBotServer, codec MenuServiceBotCodec, render func(ctx context.Context, request *telegram.Update, msg *telegram.Message) error) map[string]func(ctx context.Context, request *telegram.Update) error {
    handlers := make(map[string]func(ctx context.Context, request *telegram.Update) error)
    handlers[OperationBotMenuServiceUpdateCount] = _MenuService_UpdateCount0_Bot_Handler(srv, codec, render)
    handlers[OperationBotMenuServiceProcessMenu] = _MenuService_ProcessMenu0_Bot_Handler(srv, codec, render)
    return handlers
}
```

## Usage Examples

### Implementing the Server Interface

```go
type menuService struct {
    counter int64
}

func (s *menuService) UpdateCount(ctx context.Context, req *botv1.UpdateCountRequest) (*botv1.UpdateCountResponse, error) {
    s.counter += req.Value + req.Offset
    return &botv1.UpdateCountResponse{
        Value: s.counter,
    }, nil
}

func (s *menuService) ProcessMenu(ctx context.Context, req *botv1.ProcessMenuRequest) (*botv1.ProcessMenuResponse, error) {
    result := fmt.Sprintf("Processed menu %s with action %s", req.MenuId, req.Action)
    return &botv1.ProcessMenuResponse{
        Result: result,
    }, nil
}
```

### Implementing the Codec Interface

```go
type menuCodec struct{}

func (c *menuCodec) DecodeUpdateCountRequest(ctx context.Context, update *telegram.Update) (*botv1.UpdateCountRequest, error) {
    // Extract values from telegram update
    // Implementation depends on your protocol specifics
    value := extractValueFromUpdate(update)
    offset := extractOffsetFromUpdate(update)
    
    return &botv1.UpdateCountRequest{
        Value:  value,
        Offset: offset,
    }, nil
}

func (c *menuCodec) EncodeUpdateCountResponse(ctx context.Context, resp *botv1.UpdateCountResponse) (*telegram.Message, error) {
    return &telegram.Message{
        Text: fmt.Sprintf("Current count: %d", resp.Value),
    }, nil
}

func (c *menuCodec) DecodeProcessMenuRequest(ctx context.Context, update *telegram.Update) (*botv1.ProcessMenuRequest, error) {
    menuID, action := extractMenuDataFromUpdate(update)
    return &botv1.ProcessMenuRequest{
        MenuId: menuID,
        Action: action,
    }, nil
}

func (c *menuCodec) EncodeProcessMenuResponse(ctx context.Context, resp *botv1.ProcessMenuResponse) (*telegram.Message, error) {
    return &telegram.Message{
        Text: resp.Result,
    }, nil
}
```

### Setting Up the Router

```go
func setupBotRouter() {
    srv := &menuService{}
    codec := &menuCodec{}
    
    render := func(ctx context.Context, request *telegram.Update, msg *telegram.Message) error {
        // Send message back to Telegram
        return sendTelegramMessage(ctx, msg)
    }
    
    handlers := botv1.RegisterMenuServiceBotServer(srv, codec, render)
    
    // Register handlers with your bot framework
    for operation, handler := range handlers {
        bot.RegisterHandler(operation, handler)
    }
}
```

### Using Extra Data for Route Matching

```go
func matchRoute(update *telegram.Update) string {
    for _, operation := range botv1.GetAllBotMenuServiceOperations() {
        extraData := botv1.GetExtraBotDataByMenuServiceOperation(operation)
        if extraData != nil && matchesExtra(update, extraData) {
            return operation
        }
    }
    return ""
}

func matchesExtra(update *telegram.Update, extraData *telegram.MethodExtraData) bool {
    // Check if update matches the extra data criteria
    // Implementation depends on your matching logic
    return checkCommandMatch(update, extraData) || checkCallbackMatch(update, extraData)
}
```

## Configuration Examples

### Basic HTTP Router

```yaml
plugins:
  - local: protoc-gen-route
    out: api
    opt:
      - paths=source_relative
      - options_key=http
      - file_suffix=_http.pb.go
      - request_model=github.com/gin-gonic/gin;Context
      - response_model=net/http;ResponseWriter
```

### Telegram Bot Router

```yaml
plugins:
  - local: protoc-gen-route
    out: api
    opt:
      - paths=source_relative
      - options_key=bot
      - file_suffix=_bot.pb.go
      - request_model=github.com/go-sphere/sphere/social/telegram;Update
      - response_model=github.com/go-sphere/sphere/social/telegram;Message
      - extra_data_model=github.com/go-sphere/sphere/social/telegram;MethodExtraData
      - extra_data_constructor=github.com/go-sphere/sphere/social/telegram;NewMethodExtraData
```

### gRPC Gateway Router

```yaml
plugins:
  - local: protoc-gen-route
    out: api
    opt:
      - paths=source_relative
      - options_key=grpc
      - file_suffix=_grpc.pb.go
      - request_model=context;Context
      - response_model=google.golang.org/grpc;ServerStream
```

## Integration with Other Sphere Components

The route plugin works seamlessly with other sphere components:

- **protoc-gen-sphere**: HTTP handlers can be generated alongside route handlers
- **protoc-gen-sphere-errors**: Error handling integrates with route handlers
- **sphere/social/telegram**: Built-in support for Telegram bot routing
- **sphere/server/ginx**: HTTP routing integration with Gin framework
- **sphere/options**: Core options framework for route configuration

## Advanced Features

### Custom Templates

You can provide a custom Go template file to override the default generation:

```yaml
plugins:
  - local: protoc-gen-route
    out: api
    opt:
      - template_file=custom_route_template.go.tmpl
      - request_model=MyCustomRequest
      - response_model=MyCustomResponse
```

### Multiple Route Keys

Generate multiple route handlers for different protocols:

```yaml
plugins:
  - local: protoc-gen-route
    out: api
    opt:
      - options_key=http
      - file_suffix=_http.pb.go
      - request_model=github.com/gin-gonic/gin;Context
      - response_model=interface{}
  - local: protoc-gen-route
    out: api
    opt:
      - options_key=bot
      - file_suffix=_bot.pb.go
      - request_model=github.com/go-sphere/sphere/social/telegram;Update
      - response_model=github.com/go-sphere/sphere/social/telegram;Message
```

## Best Practices

### 1. Use Meaningful Option Keys

```protobuf
service UserService {
  rpc GetUser(GetUserRequest) returns (GetUserResponse) {
    option (sphere.options.options) = {
      key: "http"
      extra: [
        {
          key: "method"
          value: "GET"
        },
        {
          key: "path"
          value: "/users/{id}"
        }
      ]
    };
  }
}
```

### 2. Group Related Operations

```protobuf
service BotService {
  rpc HandleStart(StartRequest) returns (StartResponse) {
    option (sphere.options.options) = {
      key: "bot"
      extra: [
        {
          key: "command"
          value: "start"
        }
      ]
    };
  }
  
  rpc HandleMenu(MenuRequest) returns (MenuResponse) {
    option (sphere.options.options) = {
      key: "bot"
      extra: [
        {
          key: "callback_query"
          value: "menu_.*"
        }
      ]
    };
  }
}
```

### 3. Implement Proper Error Handling

```go
func (s *service) UpdateCount(ctx context.Context, req *botv1.UpdateCountRequest) (*botv1.UpdateCountResponse, error) {
    if req.Value < 0 {
        return nil, fmt.Errorf("invalid value: %d", req.Value)
    }
    
    // Business logic...
    
    return &botv1.UpdateCountResponse{
        Value: newValue,
    }, nil
}
```

### 4. Use Type-Safe Codec Implementations

```go
type safeCodec struct{}

func (c *safeCodec) DecodeUpdateCountRequest(ctx context.Context, update *telegram.Update) (*botv1.UpdateCountRequest, error) {
    if update == nil {
        return nil, errors.New("update is nil")
    }
    
    // Safe extraction with validation
    value, err := extractSafeValue(update)
    if err != nil {
        return nil, fmt.Errorf("failed to extract value: %w", err)
    }
    
    return &botv1.UpdateCountRequest{
        Value: value,
    }, nil
}
```

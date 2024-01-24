<p align="center">
<a href="https://www.traceloop.com/openllmetry#gh-light-mode-only">
<img width="600" src="https://raw.githubusercontent.com/traceloop/openllmetry/main/img/logo-light.png">
</a>
<a href="https://www.traceloop.com/openllmetry#gh-dark-mode-only">
<img width="600" src="https://raw.githubusercontent.com/traceloop/openllmetry/main/img/logo-dark.png">
</a>
</p>
<h1 align="center">For Go</h1>
<p align="center">
  <p align="center">Open-source observability for your LLM application</p>
</p>
<h4 align="center">
    <a href="https://traceloop.com/docs/openllmetry/getting-started-go"><strong>Get started »</strong></a>
    <br />
    <br />
  <a href="https://traceloop.com/slack">Slack</a> |
  <a href="https://traceloop.com/docs/openllmetry/introduction">Docs</a> |
  <a href="https://www.traceloop.com">Website</a>
</h4>

<h4 align="center">
   <a href="https://github.com/traceloop/go-openllmetry/blob/main/LICENSE">
    <img src="https://img.shields.io/badge/license-Apache 2.0-blue.svg" alt="OpenLLMetry is released under the Apache-2.0 License">
  </a>
  <a href="https://www.ycombinator.com/companies/traceloop"><img src="https://img.shields.io/website?color=%23f26522&down_message=Y%20Combinator&label=Backed&logo=ycombinator&style=flat-square&up_message=Y%20Combinator&url=https%3A%2F%2Fwww.ycombinator.com"></a>
  <a href="https://github.com/traceloop/go-openllmetry/blob/main/CONTRIBUTING.md">
    <img src="https://img.shields.io/badge/PRs-Welcome-brightgreen" alt="PRs welcome!" />
  </a>
  <a href="https://github.com/traceloop/go-openllmetry/issues">
    <img src="https://img.shields.io/github/commit-activity/m/traceloop/go-openllmetry" alt="git commit activity" />
  </a>
  <a href="https://traceloop.com/slack">
    <img src="https://img.shields.io/badge/chat-on%20Slack-blueviolet" alt="Slack community channel" />
  </a>
  <a href="https://twitter.com/traceloopdev">
    <img src="https://img.shields.io/badge/follow-%40traceloopdev-1DA1F2?logo=twitter&style=social" alt="Traceloop Twitter" />
  </a>
</h4>

OpenLLMetry is a set of extensions built on top of [OpenTelemetry](https://opentelemetry.io/) that gives you complete observability over your LLM application. Because it uses OpenTelemetry under the hood, it can be connected to your existing observability solutions - Datadog, Honeycomb, and others.

It's built and maintained by Traceloop under the Apache 2.0 license.

The repo contains standard OpenTelemetry instrumentations for LLM providers and Vector DBs, as well as a Traceloop SDK that makes it easy to get started with OpenLLMetry, while still outputting standard OpenTelemetry data that can be connected to your observability stack.
If you already have OpenTelemetry instrumented, you can just add any of our instrumentations directly.

## 🚀 Getting Started

The easiest way to get started is to use our SDK.
For a complete guide, go to our [docs](https://traceloop.com/docs/openllmetry/getting-started-go).

Install the SDK:

```bash
go get traceloop-sdk
```

Then, initialize the SDK in your code:

```go
package main

import (
	"context"

	sdk "github.com/traceloop/go-openllmetry/traceloop-sdk"
	"github.com/traceloop/go-openllmetry/traceloop-sdk/config"
)

func main() {
    ctx := context.Background()

    traceloop := sdk.NewClient(ctx, config.Config{
		APIKey: os.Getenv("TRACELOOP_API_KEY"),
	})
	defer func() { traceloop.Shutdown(ctx) }()
}
```

That's it. You're now tracing your code with OpenLLMetry!

Now, you need to decide where to export the traces to.

## ⏫ Supported (and tested) destinations

- [x] [Traceloop](https://www.traceloop.com/docs/openllmetry/integrations/traceloop)
- [x] [Dynatrace](https://www.traceloop.com/docs/openllmetry/integrations/dynatrace)
- [x] [Datadog](https://www.traceloop.com/docs/openllmetry/integrations/datadog)
- [x] [New Relic](https://www.traceloop.com/docs/openllmetry/integrations/newrelic)
- [x] [Honeycomb](https://www.traceloop.com/docs/openllmetry/integrations/honeycomb)
- [x] [Grafana Tempo](https://www.traceloop.com/docs/openllmetry/integrations/grafana)
- [x] [HyperDX](https://www.traceloop.com/docs/openllmetry/integrations/hyperdx)
- [x] [SigNoz](https://www.traceloop.com/docs/openllmetry/integrations/signoz)
- [x] [OpenTelemetry Collector](https://www.traceloop.com/docs/openllmetry/integrations/otel-collector)

See [our docs](https://traceloop.com/docs/openllmetry/integrations/exporting) for instructions on connecting to each one.

## 🪗 What do we instrument?

OpenLLMetry is in early-alpha exploratory stage, and we're still figuring out what to instrument.
As opposed to other languages, there aren't many official LLM libraries (yet?), so for now you'll have to manually log prompts:

```go
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/sashabaranov/go-openai"
	sdk "github.com/traceloop/go-openllmetry/traceloop-sdk"
	"github.com/traceloop/go-openllmetry/traceloop-sdk/config"
	"github.com/traceloop/go-openllmetry/traceloop-sdk/dto"
)

func main() {
	ctx := context.Background()

	// Initialize Traceloop
	traceloop := sdk.NewClient(ctx, config.Config{
		APIKey:  os.Getenv("TRACELOOP_API_KEY"),
	})
	defer func() { traceloop.Shutdown(ctx) }()

	// Call OpenAI like you normally would
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: "Tell me a joke about OpenTelemetry!",
				},
			},
		},
	)

	// Log the request and the response
	log := dto.PromptLogAttributes{
		Prompt: dto.Prompt{
			Vendor: "openai",
			Mode:   "chat",
			Model:  request.Model,
		},
		Completion: dto.Completion{
			Model: resp.Model,
		},
		Usage: dto.Usage{
			TotalTokens:      resp.Usage.TotalTokens,
			CompletionTokens: resp.Usage.CompletionTokens,
			PromptTokens:     resp.Usage.PromptTokens,
		},
	}

	for i, message := range request.Messages {
		log.Prompt.Messages = append(log.Prompt.Messages, dto.Message{
			Index:   i,
			Content: message.Content,
			Role:    message.Role,
		})
	}

	for _, choice := range resp.Choices {
		log.Completion.Messages = append(log.Completion.Messages, dto.Message{
			Index:   choice.Index,
			Content: choice.Message.Content,
			Role:    choice.Message.Role,
		})
	}

	traceloop.LogPrompt(ctx, log)
}
```

## 🌱 Contributing

Whether it's big or small, we love contributions ❤️ Check out our guide to see how to [get started](https://traceloop.com/docs/openllmetry/contributing/overview).

Not sure where to get started? You can:

- [Book a free pairing session with one of our teammates](mailto:nir@traceloop.com?subject=Pairing%20session&body=I'd%20like%20to%20do%20a%20pairing%20session!)!
- Join our <a href="https://join.slack.com/t/traceloopcommunity/shared_invite/zt-1plpfpm6r-zOHKI028VkpcWdobX65C~g">Slack</a>, and ask us any questions there.

## 💚 Community & Support

- [Slack](https://join.slack.com/t/traceloopcommunity/shared_invite/zt-1plpfpm6r-zOHKI028VkpcWdobX65C~g) (For live discussion with the community and the Traceloop team)
- [GitHub Discussions](https://github.com/traceloop/go-openllmetry/discussions) (For help with building and deeper conversations about features)
- [GitHub Issues](https://github.com/traceloop/go-openllmetry/issues) (For any bugs and errors you encounter using OpenLLMetry)
- [Twitter](https://twitter.com/traceloopdev) (Get news fast)

## 🙏 Special Thanks

To @patrickdebois, who [suggested the great name](https://x.com/patrickdebois/status/1695518950715473991?s=46&t=zn2SOuJcSVq-Pe2Ysevzkg) we're now using for this repo!

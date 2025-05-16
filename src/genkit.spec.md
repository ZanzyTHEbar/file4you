Here is the API specification for `github.com/firebase/genkit/go` and its key sub-packages, based on the successfully retrieved `go doc` output and inferences:

```markdown
# API Specification: github.com/firebase/genkit/go

## Package: github.com/firebase/genkit/go/genkit

Package genkit provides Genkit functionality for application developers.

### Types

*   **`Genkit` struct**:
    *   `func Init(ctx context.Context, opts ...GenkitOption) (*Genkit, error)`: Initializes a new Genkit instance.
*   **`GenkitOption` interface**: Interface for options to configure Genkit.
    *   `func WithDefaultModel(model string) GenkitOption`: Sets the default model.
    *   `func WithPlugins(plugins ...Plugin) GenkitOption`: Registers plugins.
    *   `func WithPromptDir(dir string) GenkitOption`: Sets the directory for loading prompts.
*   **`HandlerOption` interface**: Interface for options to configure HTTP handlers.
    *   `func WithContextProviders(ctxProviders ...core.ContextProvider) HandlerOption`: Adds context providers to the handler.
*   **`Plugin` interface**: Interface for Genkit plugins. (Details likely in `core` or a dedicated `plugin` package).

### Functions

*   `func DefineBatchEvaluator(g *Genkit, provider, name string, options *ai.EvaluatorOptions, ...) (ai.Evaluator, error)`: Defines a batch evaluator.
*   `func DefineEmbedder(g *Genkit, provider, name string, ...) ai.Embedder`: Defines an embedder.
*   `func DefineEvaluator(g *Genkit, provider, name string, options *ai.EvaluatorOptions, ...) (ai.Evaluator, error)`: Defines an evaluator.
*   `func DefineFlow[In, Out any](g *Genkit, name string, fn core.Func[In, Out]) *core.Flow[In, Out, struct{}]`: Defines a regular flow.
*   `func DefineFormat(g *Genkit, name string, formatter ai.Formatter)`: Defines a custom output formatter.
*   `func DefineHelper(g *Genkit, name string, fn any) error`: Defines a helper function.
*   `func DefineIndexer(g *Genkit, provider, name string, ...) ai.Indexer`: Defines an indexer.
*   `func DefineModel(g *Genkit, provider, name string, info *ai.ModelInfo, fn ai.ModelFunc) ai.Model`: Defines a model.
*   `func DefinePartial(g *Genkit, name string, source string) error`: Defines a partial template.
*   `func DefinePrompt(g *Genkit, name string, opts ...ai.PromptOption) (*ai.Prompt, error)`: Defines a prompt.
*   `func DefineRetriever(g *Genkit, provider, name string, ...) ai.Retriever`: Defines a retriever.
*   `func DefineStreamingFlow[In, Out, Stream any](g *Genkit, name string, fn core.StreamingFunc[In, Out, Stream]) *core.Flow[In, Out, Stream]`: Defines a streaming flow.
*   `func DefineTool[In, Out any](g *Genkit, name, description string, ...) ai.Tool`: Defines a tool (action).
*   `func Generate(ctx context.Context, g *Genkit, opts ...ai.GenerateOption) (*ai.ModelResponse, error)`: Performs generation with a model.
*   `func GenerateData[Out any](ctx context.Context, g *Genkit, opts ...ai.GenerateOption) (*Out, *ai.ModelResponse, error)`: Performs generation and unmarshals structured data.
*   `func GenerateText(ctx context.Context, g *Genkit, opts ...ai.GenerateOption) (string, error)`: Performs generation and returns text.
*   `func GenerateWithRequest(ctx context.Context, g *Genkit, actionOpts *ai.GenerateActionOptions, ...) (*ai.ModelResponse, error)`: Performs generation with more detailed request options.
*   `func Handler(a core.Action, opts ...HandlerOption) http.HandlerFunc`: Creates an HTTP handler for an action.
*   `func IsDefinedFormat(g *Genkit, name string) bool`: Checks if a format is defined.
*   `func ListFlows(g *Genkit) []core.Action`: Lists all defined flows.
*   `func LoadPrompt(g *Genkit, path string, namespace string) (*ai.Prompt, error)`: Loads a prompt from a file.
*   `func LoadPromptDir(g *Genkit, dir string, namespace string) error`: Loads all prompts from a directory.
*   `func LookupEmbedder(g *Genkit, provider, name string) ai.Embedder`: Looks up a defined embedder.
*   `func LookupEvaluator(g *Genkit, provider, name string) ai.Evaluator`: Looks up a defined evaluator.
*   `func LookupIndexer(g *Genkit, provider, name string) ai.Indexer`: Looks up a defined indexer.
*   `func LookupModel(g *Genkit, provider, name string) ai.Model`: Looks up a defined model.
*   `func LookupPlugin(g *Genkit, name string) any`: Looks up a registered plugin.
*   `func LookupPrompt(g *Genkit, name string) *ai.Prompt`: Looks up a defined prompt.
*   `func LookupRetriever(g *Genkit, provider, name string) ai.Retriever`: Looks up a defined retriever.
*   `func LookupTool(g *Genkit, name string) ai.Tool`: Looks up a defined tool.
*   `func RegisterSpanProcessor(g *Genkit, sp sdktrace.SpanProcessor)`: Registers a span processor for telemetry.
*   `func Run[Out any](ctx context.Context, name string, fn func() (Out, error)) (Out, error)`: Executes a function with tracing. (Likely delegates to `core.Run`).

## Package: github.com/firebase/genkit/go/core

Package core implements Genkit actions and other essential machinery. Intended for Genkit internals and plugins.

### Constants

*   `const CodeOK = 0`: Represents successful status. (Other status codes likely exist).

### Variables

*   `var StatusNameToCode = map[StatusName]int{ ... }`: Maps status names to integer codes.

### Types

*   **`Action` interface**: Represents a runnable Genkit action.
*   **`ActionContext` type**: `map[string]any`. Stores context for an action.
    *   `func FromContext(ctx context.Context) ActionContext`: Extracts ActionContext from a Go context.
*   **`ActionDef[In, Out, Stream any]` struct**: Definition of an action.
    *   `func DefineAction[In, Out any](r *registry.Registry, provider, name string, atype atype.ActionType, ...) *ActionDef[In, Out, struct{}]`: Defines a non-streaming action.
    *   `func DefineActionWithInputSchema[Out any](r *registry.Registry, provider, name string, atype atype.ActionType, ...) *ActionDef[any, Out, struct{}]`: Defines an action with an input schema.
    *   `func DefineStreamingAction[In, Out, Stream any](r *registry.Registry, provider, name string, atype atype.ActionType, ...) *ActionDef[In, Out, Stream]`: Defines a streaming action.
    *   `func LookupActionFor[In, Out, Stream any](r *registry.Registry, typ atype.ActionType, provider, name string) *ActionDef[In, Out, Stream]`: Looks up an action definition.
*   **`ContextProvider` type**: `func(ctx context.Context, req RequestData) (ActionContext, error)`. Function type for providing action context.
*   **`Flow[In, Out, Stream any]` type**: Alias for `ActionDef[In, Out, Stream]`. Represents a flow.
    *   `func DefineFlow[In, Out any](r *registry.Registry, name string, fn Func[In, Out]) *Flow[In, Out, struct{}]`: Defines a non-streaming flow.
    *   `func DefineStreamingFlow[In, Out, Stream any](r *registry.Registry, name string, fn StreamingFunc[In, Out, Stream]) *Flow[In, Out, Stream]`: Defines a streaming flow.
*   **`Func[In, Out any]` type**: `func(context.Context, In) (Out, error)`. Signature for a non-streaming action/flow implementation.
*   **`GenkitError` struct**: Represents an error within Genkit.
    *   `func NewError(status StatusName, message string, args ...any) *GenkitError`: Creates a new GenkitError.
*   **`Middleware[In, Out, Stream any]` type**: `func(StreamingFunc[In, Out, Stream]) StreamingFunc[In, Out, Stream]`. Signature for action/flow middleware.
    *   `func ChainMiddleware[In, Out, Stream any](middlewares ...Middleware[In, Out, Stream]) Middleware[In, Out, Stream]`: Chains multiple middlewares.
    *   `func Middlewares[In, Out, Stream any](ms ...Middleware[In, Out, Stream]) []Middleware[In, Out, Stream]`: Returns a slice of middlewares.
*   **`ReflectionError` struct**: Error related to reflection operations.
    *   `func ToReflectionError(err error) ReflectionError`: Converts a standard error to ReflectionError.
*   **`ReflectionErrorDetails` struct**: Details for a reflection error.
*   **`RequestData` struct**: Represents data associated with a request.
*   **`Status` struct**: Represents a status, including code and message.
    *   `func NewStatus(name StatusName, message string) *Status`: Creates a new Status.
*   **`StatusName` string**: Type for canonical status names (e.g., "OK", "CANCELLED").
    *   `const OK StatusName = "OK"` (and other status names).
*   **`StreamCallback[Stream any]` type**: `func(context.Context, Stream) error`. Callback for handling streamed data.
*   **`StreamingFlowValue[Out, Stream any]` struct**: Value returned by a streaming flow, containing output and stream.
*   **`StreamingFunc[In, Out, Stream any]` type**: `func(context.Context, In, StreamCallback[Stream]) (Out, error)`. Signature for a streaming action/flow implementation.
*   **`UserFacingError` struct**: Error intended to be shown to end-users.
    *   `func NewPublicError(status StatusName, message string, details map[string]any) *UserFacingError`: Creates a new UserFacingError.

### Functions

*   `func HTTPStatusCode(name StatusName) int`: Returns the corresponding HTTP status code for a Genkit StatusName.
*   `func RegisterSpanProcessor(r *registry.Registry, sp sdktrace.SpanProcessor)`: Registers a span processor with a registry.
*   `func Run[Out any](ctx context.Context, name string, fn func() (Out, error)) (Out, error)`: Executes a named function with tracing and error handling.
*   `func WithActionContext(ctx context.Context, actionCtx ActionContext) context.Context`: Adds ActionContext to a Go context.

## Package: github.com/firebase/genkit/go/ai (Inferred)

This package likely defines interfaces and structs related to AI components like models, tools, prompts, etc. The following are inferred from their usage in the `genkit` package:

### Likely Types (Interfaces or Structs)

*   **`ai.EvaluatorOptions`**: Options for configuring an evaluator.
*   **`ai.Evaluator`**: Interface or struct for an evaluator.
*   **`ai.Embedder`**: Interface or struct for an embedder.
*   **`ai.Formatter`**: Interface for custom output formatters.
*   **`ai.Indexer`**: Interface or struct for an indexer.
*   **`ai.ModelInfo`**: Struct containing metadata about a model.
*   **`ai.ModelFunc`**: Function signature for a model's implementation.
*   **`ai.Model`**: Interface or struct representing a model.
*   **`ai.PromptOption`**: Options for configuring a prompt.
*   **`ai.Prompt`**: Struct representing a defined prompt.
*   **`ai.Retriever`**: Interface or struct for a retriever.
*   **`ai.Tool`**: Interface or struct representing a tool/action.
*   **`ai.GenerateOption`**: Options for the `Generate` function.
*   **`ai.ModelResponse`**: Struct representing the response from a model.
*   **`ai.GenerateActionOptions`**: Detailed options for a generate action.

This specification should provide a good overview of the public API for `github.com/firebase/genkit/go`.

---

Directories:

ai
core: Package core implements Genkit actions and other essential machinery.
    - logger: Package logger provides a context-scoped slog.Logger.
    - tracing: Package gtime provides time functionality for Go Genkit.
genkit: Package genkit provides Genkit functionality for application developers.
plugins
    - evaluators: Package evaluators defines a set of Genkit Evaluators for popular use-cases
    - firebase
    - googlecloud: The googlecloud package supports telemetry (tracing, metrics and logging) using Google Cloud services.
    - googlegenai
    - localvec: Package localvec is a local vector database for development and testing.
    - ollama
    - pinecone: Package pinecone implements a genkit plugin for the Pinecone vector database.
    - server
    - vertexai/modelgarden
    - weaviate
samples
    - basic-gemini
    - basic-gemini-with-context
    - cache-gemini
    - code-execution-gemini
    - coffee-shop
    - firebase-retrievers
    - flow-sample1
    - formats
    - imagen-gemini
    - menu
    - modelgarden
    - ollama-tools
    - ollama-vision
    - partials-and-helpers
    - pgvector
    - prompts
    - prompts-dir: [START main]
    - rag
utils

---

Tool calling 

bookmark_border
Tool calling, also known as function calling, is a structured way to give LLMs the ability to make requests back to the application that called it. You define the tools you want to make available to the model, and the model will make tool requests to your app as necessary to fulfill the prompts you give it.

The use cases of tool calling generally fall into a few themes:

Giving an LLM access to information it wasn't trained with

Frequently changing information, such as a stock price or the current weather.
Information specific to your app domain, such as product information or user profiles.
Note the overlap with retrieval augmented generation (RAG), which is also a way to let an LLM integrate factual information into its generations. RAG is a heavier solution that is most suited when you have a large amount of information or the information that's most relevant to a prompt is ambiguous. On the other hand, if a function call or database lookup is all that's necessary for retrieving the information the LLM needs, tool calling is more appropriate.

Introducing a degree of determinism into an LLM workflow

Performing calculations that the LLM cannot reliably complete itself.
Forcing an LLM to generate verbatim text under certain circumstances, such as when responding to a question about an app's terms of service.
Performing an action when initiated by an LLM

Turning on and off lights in an LLM-powered home assistant
Reserving table reservations in an LLM-powered restaurant agent
Before you begin
If you want to run the code examples on this page, first complete the steps in the Get started guide. All of the examples assume that you have already set up a project with Genkit dependencies installed.

This page discusses one of the advanced features of Genkit model abstraction, so before you dive too deeply, you should be familiar with the content on the Generating content with AI models page. You should also be familiar with Genkit's system for defining input and output schemas, which is discussed on the Flows page.

Overview of tool calling
At a high level, this is what a typical tool-calling interaction with an LLM looks like:

The calling application prompts the LLM with a request and also includes in the prompt a list of tools the LLM can use to generate a response.
The LLM either generates a complete response or generates a tool call request in a specific format.
If the caller receives a complete response, the request is fulfilled and the interaction ends; but if the caller receives a tool call, it performs whatever logic is appropriate and sends a new request to the LLM containing the original prompt or some variation of it as well as the result of the tool call.
The LLM handles the new prompt as in Step 2.
For this to work, several requirements must be met:

The model must be trained to make tool requests when it's needed to complete a prompt. Most of the larger models provided through web APIs such as Gemini can do this, but smaller and more specialized models often cannot. Genkit will throw an error if you try to provide tools to a model that doesn't support it.
The calling application must provide tool definitions to the model in the format it expects.
The calling application must prompt the model to generate tool calling requests in the format the application expects.
Tool calling with Genkit
Genkit provides a single interface for tool calling with models that support it. Each model plugin ensures that the last two criteria mentioned in the previous section are met, and the genkit.Generate() function automatically carries out the tool-calling loop described earlier.

Model support
Tool calling support depends on the model, the model API, and the Genkit plugin. Consult the relevant documentation to determine if tool calling is likely to be supported. In addition:

Genkit will throw an error if you try to provide tools to a model that doesn't support it.
If the plugin exports model references, the ModelInfo.Supports.Tools property will indicate if it supports tool calling.
Defining tools
Use the genkit.DefineTool() function to write tool definitions:


import (
    "context"
    "log"

    "github.com/firebase/genkit/go/ai"
    "github.com/firebase/genkit/go/genkit"
    "github.com/firebase/genkit/go/plugins/googlegenai"
)

func main() {
    ctx := context.Background()

    g, err := genkit.Init(ctx,
        genkit.WithPlugins(&googlegenai.GoogleAI{}),
        genkit.WithDefaultModel("googleai/gemini-2.0-flash"),
    )
    if err != nil {
      log.Fatal(err)
    }

    getWeatherTool := genkit.DefineTool(
        g, "getWeather", "Gets the current weather in a given location",
        func(ctx *ai.ToolContext, input struct{
            Location string `jsonschema_description:"Location to get weather for"`
        }) (string, error) {
            // Here, we would typically make an API call or database query. For this
            // example, we just return a fixed value.
            return fmt.Sprintf("The current weather in %s is 63°F and sunny.", input.Location);
        })
}
The syntax here looks just like the genkit.DefineFlow() syntax; however, you must write a description. Take special care with the wording and descriptiveness of the description as it is vital for the LLM to decide to use it appropriately.

Using tools
Include defined tools in your prompts to generate content.

Generate
DefinePrompt
Prompt file

resp, err := genkit.Generate(ctx, g,
    ai.WithPrompt("What is the weather in San Francisco?"),
    ai.WithTools(getWeatherTool),
)
Genkit will automatically handle the tool call if the LLM needs to use the getWeather tool to answer the prompt.

Explicitly handling tool calls
If you want full control over this tool-calling loop, for example to apply more complicated logic, set the WithReturnToolRequests() option to true. Now it's your responsibility to ensure all of the tool requests are fulfilled:


getWeatherTool := genkit.DefineTool(
    g, "getWeather", "Gets the current weather in a given location",
    func(ctx *ai.ToolContext, location struct {
        Location string `jsonschema_description:"Location to get weather for"`
    }) (string, error) {
        // Tool implementation...
        return "sunny", nil
    },
)

resp, err := genkit.Generate(ctx, g,
    ai.WithPrompt("What is the weather in San Francisco?"),
    ai.WithTools(getWeatherTool),
    ai.WithReturnToolRequests(true),
)
if err != nil {
    log.Fatal(err)
}

parts := []*ai.Part{}
for _, req := range resp.ToolRequests() {
    tool := genkit.LookupTool(g, req.Name)
    if tool == nil {
        log.Fatalf("tool %q not found", req.Name)
    }

    output, err := tool.RunRaw(ctx, req.Input)
    if err != nil {
        log.Fatalf("tool %q execution failed: %v", tool.Name(), err)
    }

    parts = append(parts,
        ai.NewToolResponsePart(&ai.ToolResponse{
            Name:   req.Name,
            Ref:    req.Ref,
            Output: output,
        }))
}

resp, err = genkit.Generate(ctx, g,
    ai.WithMessages(append(resp.History(), ai.NewMessage(ai.RoleTool, nil, parts...))...),
)
if err != nil {
    log.Fatal(err)
}

---

At the heart of generative AI are AI models. The two most prominent examples of generative models are large language models (LLMs) and image generation models. These models take input, called a prompt (most commonly text, an image, or a combination of both), and from it produce as output text, an image, or even audio or video.

The output of these models can be surprisingly convincing: LLMs generate text that appears as though it could have been written by a human being, and image generation models can produce images that are very close to real photographs or artwork created by humans.

In addition, LLMs have proven capable of tasks beyond simple text generation:

Writing computer programs.
Planning subtasks that are required to complete a larger task.
Organizing unorganized data.
Understanding and extracting information data from a corpus of text.
Following and performing automated activities based on a text description of the activity.
There are many models available to you, from several different providers. Each model has its own strengths and weaknesses and one model might excel at one task but perform less well at others. Apps making use of generative AI can often benefit from using multiple different models depending on the task at hand.

As an app developer, you typically don't interact with generative AI models directly, but rather through services available as web APIs. Although these services often have similar functionality, they all provide them through different and incompatible APIs. If you want to make use of multiple model services, you have to use each of their proprietary SDKs, potentially incompatible with each other. And if you want to upgrade from one model to the newest and most capable one, you might have to build that integration all over again.

Genkit addresses this challenge by providing a single interface that abstracts away the details of accessing potentially any generative AI model service, with several prebuilt implementations already available. Building your AI-powered app around Genkit simplifies the process of making your first generative AI call and makes it equally straightforward to combine multiple models or swap one model for another as new models emerge.

Before you begin
If you want to run the code examples on this page, first complete the steps in the Get started guide. All of the examples assume that you have already installed Genkit as a dependency in your project.

Models supported by Genkit
Genkit is designed to be flexible enough to use potentially any generative AI model service. Its core libraries define the common interface for working with models, and model plugins define the implementation details for working with a specific model and its API.

The Genkit team maintains plugins for working with models provided by Vertex AI, Google Generative AI, and Ollama:

Gemini family of LLMs, through the Google Cloud Vertex AI plugin.
Gemini family of LLMs, through the Google AI plugin.
Gemma 3, Llama 4, and many more open models, through the Ollama plugin (you must host the Ollama server yourself).
Loading and configuring model plugins
Before you can use Genkit to start generating content, you need to load and configure a model plugin. If you're coming from the Get Started guide, you've already done this. Otherwise, see the Get Started guide or the individual plugin's documentation and follow the steps there before continuing.

The genkit.Generate() function
In Genkit, the primary interface through which you interact with generative AI models is the genkit.Generate() function.

The simplest genkit.Generate() call specifies the model you want to use and a text prompt:


package main

import (
    "context"
    "log"

    "github.com/firebase/genkit/go/ai"
    "github.com/firebase/genkit/go/genkit"
    "github.com/firebase/genkit/go/plugins/googlegenai"
)

func main() {
    ctx := context.Background()

    g, err := genkit.Init(ctx,
        genkit.WithPlugins(&googlegenai.GoogleAI{}),
        genkit.WithDefaultModel("googleai/gemini-2.0-flash"),
    )
    if err != nil {
        log.Fatal("could not initialize Genkit: %w", err)
    }

    resp, err := genkit.Generate(ctx, g,
        ai.WithPrompt("Invent a menu item for a pirate themed restaurant."),
    )
    if err != nil {
        log.Fatal("could not generate model response: %w", err)
    }

    log.Println(resp.Text())
}
When you run this brief example, it will print out some debugging information followed by the output of the genkit.Generate() call, which will usually be Markdown text as in the following example:


## The Blackheart's Bounty

**A hearty stew of slow-cooked beef, spiced with rum and molasses, served in a
hollowed-out cannonball with a side of crusty bread and a dollop of tangy
pineapple salsa.**

**Description:** This dish is a tribute to the hearty meals enjoyed by pirates
on the high seas. The beef is tender and flavorful, infused with the warm spices
of rum and molasses. The pineapple salsa adds a touch of sweetness and acidity,
balancing the richness of the stew. The cannonball serving vessel adds a fun and
thematic touch, making this dish a perfect choice for any pirate-themed
adventure.
Run the script again and you'll get a different output.

The preceding code sample sent the generation request to the default model, which you specified when you configured the Genkit instance.

You can also specify a model for a single genkit.Generate() call:


resp, err := genkit.Generate(ctx, g,
    ai.WithModelName("googleai/gemini-2.5-pro"),
    ai.WithPrompt("Invent a menu item for a pirate themed restaurant."),
)
A model string identifier looks like providerid/modelid, where the provider ID (in this case, googleai) identifies the plugin, and the model ID is a plugin-specific string identifier for a specific version of a model.

These examples also illustrate an important point: when you use genkit.Generate() to make generative AI model calls, changing the model you want to use is a matter of passing a different value to the model parameter. By using genkit.Generate() instead of the native model SDKs, you give yourself the flexibility to more easily use several different models in your app and change models in the future.

So far you have only seen examples of the simplest genkit.Generate() calls. However, genkit.Generate() also provides an interface for more advanced interactions with generative models, which you will see in the sections that follow.

System prompts
Some models support providing a system prompt, which gives the model instructions as to how you want it to respond to messages from the user. You can use the system prompt to specify characteristics such as a persona you want the model to adopt, the tone of its responses, and the format of its responses.

If the model you're using supports system prompts, you can provide one with the WithSystem() option:


resp, err := genkit.Generate(ctx, g,
    ai.WithSystem("You are a food industry marketing consultant."),
    ai.WithPrompt("Invent a menu item for a pirate themed restaurant."),
)
For models that don't support system prompts, WithSystem() simulates it by modifying the request to appear like a system prompt.

Model parameters
The genkit.Generate() function takes a WithConfig() option, through which you can specify optional settings that control how the model generates content:


resp, err := genkit.Generate(ctx, g,
    ai.WithModelName("googleai/gemini-2.0-flash"),
    ai.WithPrompt("Invent a menu item for a pirate themed restaurant."),
    ai.WithConfig(&googlegenai.GeminiConfig{
        MaxOutputTokens: 500,
        StopSequences:   ["<end>", "<fin>"],
        Temperature:     0.5,
        TopP:            0.4,
        TopK:            50,
    }),
)
The exact parameters that are supported depend on the individual model and model API. However, the parameters in the previous example are common to almost every model. The following is an explanation of these parameters:

Parameters that control output length
MaxOutputTokens

LLMs operate on units called tokens. A token usually, but does not necessarily, map to a specific sequence of characters. When you pass a prompt to a model, one of the first steps it takes is to tokenize your prompt string into a sequence of tokens. Then, the LLM generates a sequence of tokens from the tokenized input. Finally, the sequence of tokens gets converted back into text, which is your output.

The maximum output tokens parameter sets a limit on how many tokens to generate using the LLM. Every model potentially uses a different tokenizer, but a good rule of thumb is to consider a single English word to be made of 2 to 4 tokens.

As stated earlier, some tokens might not map to character sequences. One such example is that there is often a token that indicates the end of the sequence: when an LLM generates this token, it stops generating more. Therefore, it's possible and often the case that an LLM generates fewer tokens than the maximum because it generated the "stop" token.

StopSequences

You can use this parameter to set the tokens or token sequences that, when generated, indicate the end of LLM output. The correct values to use here generally depend on how the model was trained, and are usually set by the model plugin. However, if you have prompted the model to generate another stop sequence, you might specify it here.

Note that you are specifying character sequences, and not tokens per se. In most cases, you will specify a character sequence that the model's tokenizer maps to a single token.

Parameters that control "creativity"
The temperature, top-p, and top-k parameters together control how "creative" you want the model to be. This section provides very brief explanations of what these parameters mean, but the more important point is this: these parameters are used to adjust the character of an LLM's output. The optimal values for them depend on your goals and preferences, and are likely to be found only through experimentation.

Temperature

LLMs are fundamentally token-predicting machines. For a given sequence of tokens (such as the prompt) an LLM predicts, for each token in its vocabulary, the likelihood that the token comes next in the sequence. The temperature is a scaling factor by which these predictions are divided before being normalized to a probability between 0 and 1.

Low temperature values—between 0.0 and 1.0—amplify the difference in likelihoods between tokens, with the result that the model will be even less likely to produce a token it already evaluated to be unlikely. This is often perceived as output that is less creative. Although 0.0 is technically not a valid value, many models treat it as indicating that the model should behave deterministically, and to only consider the single most likely token.

High temperature values—those greater than 1.0—compress the differences in likelihoods between tokens, with the result that the model becomes more likely to produce tokens it had previously evaluated to be unlikely. This is often perceived as output that is more creative. Some model APIs impose a maximum temperature, often 2.0.

TopP

Top-p is a value between 0.0 and 1.0 that controls the number of possible tokens you want the model to consider, by specifying the cumulative probability of the tokens. For example, a value of 1.0 means to consider every possible token (but still take into account the probability of each token). A value of 0.4 means to only consider the most likely tokens, whose probabilities add up to 0.4, and to exclude the remaining tokens from consideration.

TopK

Top-k is an integer value that also controls the number of possible tokens you want the model to consider, but this time by explicitly specifying the maximum number of tokens. Specifying a value of 1 means that the model should behave deterministically.

Experiment with model parameters
You can experiment with the effect of these parameters on the output generated by different model and prompt combinations by using the Developer UI. Start the developer UI with the genkit start command and it will automatically load all of the models defined by the plugins configured in your project. You can quickly try different prompts and configuration values without having to repeatedly make these changes in code.

Pair model with its config
Given that each provider or even a specific model may have its own configuration schema or warrant certain settings, it may be error prone to set separate options using WithModelName() and WithConfig() since the latter is not strongly typed to the former.

To pair a model with its config, you can create a model reference that you can pass into the generate call instead:


model := googlegenai.GoogleAIModelRef("gemini-2.0-flash", &googlegenai.GeminiConfig{
    MaxOutputTokens: 500,
    StopSequences:   ["<end>", "<fin>"],
    Temperature:     0.5,
    TopP:            0.4,
    TopK:            50,
})

resp, err := genkit.Generate(ctx, g,
    ai.WithModel(model),
    ai.WithPrompt("Invent a menu item for a pirate themed restaurant."),
)
if err != nil {
    log.Fatal(err)
}
The constructor for the model reference will enforce that the correct config type is provided which may reduce mismatches.

Structured output
When using generative AI as a component in your application, you often want output in a format other than plain text. Even if you're just generating content to display to the user, you can benefit from structured output simply for the purpose of presenting it more attractively to the user. But for more advanced applications of generative AI, such as programmatic use of the model's output, or feeding the output of one model into another, structured output is a must.

In Genkit, you can request structured output from a model by specifying an output type when you call genkit.Generate():


type MenuItem struct {
    Name        string   `json:"name"`
    Description string   `json:"description"`
    Calories    int      `json:"calories"`
    Allergens   []string `json:"allergens"`
}

resp, err := genkit.Generate(ctx, g,
    ai.WithPrompt("Invent a menu item for a pirate themed restaurant."),
    ai.WithOutputType(MenuItem{}),
)
if err != nil {
  log.Fatal(err) // One possible error is that the response does not conform to the type.
}
Model output types are specified as JSON schema using the invopop/jsonschema package. This provides runtime type checking, which bridges the gap between static Go types and the unpredictable output of generative AI models. This system lets you write code that can rely on the fact that a successful generate call will always return output that conforms to your Go types.

When you specify an output type in genkit.Generate(), Genkit does several things behind the scenes:

Augments the prompt with additional guidance about the selected output format. This also has the side effect of specifying to the model what content exactly you want to generate (for example, not only suggest a menu item but also generate a description, a list of allergens, and so on).
Verifies that the output conforms to the schema.
Marshals the model output into a Go type.
To get structured output from a successful generate call, call Output() on the model response with an empty value of the type:


var item MenuItem
if err := resp.Output(&item); err != nil {
    log.Fatalf(err)
}

log.Printf("%s (%d calories, %d allergens): %s\n",
    item.Name, item.Calories, len(item.Allergens), item.Description)
Alternatively, you can use genkit.GenerateData() for a more succinct call:


item, resp, err := genkit.GenerateData[MenuItem](ctx, g,
    ai.WithPrompt("Invent a menu item for a pirate themed restaurant."),
)
if err != nil {
  log.Fatal(err)
}

log.Printf("%s (%d calories, %d allergens): %s\n",
    item.Name, item.Calories, len(item.Allergens), item.Description)
This function requires the output type parameter but automatically sets the WithOutputType() option and calls resp.Output() before returning the value.

Handling errors
Note in the prior example that the genkit.Generate() call can result in an error. One possible error can happen when the model fails to generate output that conforms to the schema. The best strategy for dealing with such errors will depend on your exact use case, but here are some general hints:

Try a different model. For structured output to succeed, the model must be capable of generating output in JSON. The most powerful LLMs like Gemini are versatile enough to do this; however, smaller models, such as some of the local models you would use with Ollama, might not be able to generate structured output reliably unless they have been specifically trained to do so.

Simplify the schema. LLMs may have trouble generating complex or deeply nested types. Try using clear names, fewer fields, or a flattened structure if you are not able to reliably generate structured data.

Retry the genkit.Generate() call. If the model you've chosen only rarely fails to generate conformant output, you can treat the error as you would treat a network error, and retry the request using some kind of incremental back-off strategy.

Streaming
When generating large amounts of text, you can improve the experience for your users by presenting the output as it's generated—streaming the output. A familiar example of streaming in action can be seen in most LLM chat apps: users can read the model's response to their message as it's being generated, which improves the perceived responsiveness of the application and enhances the illusion of chatting with an intelligent counterpart.

In Genkit, you can stream output using the WithStreaming() option:


resp, err := genkit.Generate(ctx, g,
    ai.WithPrompt("Suggest a complete menu for a pirate themed restaurant."),
    ai.WithStreaming(func(ctx context.Context, chunk *ai.ModelResponseChunk) error {
        // Do something with the chunk...
        log.Println(chunk.Text())
        return nil
    }),
)
if err != nil {
    log.Fatal(err)
}

log.Println(resp.Text())
Multimodal input
The examples you've seen so far have used text strings as model prompts. While this remains the most common way to prompt generative AI models, many models can also accept other media as prompts. Media prompts are most often used in conjunction with text prompts that instruct the model to perform some operation on the media, such as to caption an image or transcribe an audio recording.

The ability to accept media input and the types of media you can use are completely dependent on the model and its API. For example, the Gemini 2.0 series of models can accept images, video, and audio as prompts.

To provide a media prompt to a model that supports it, instead of passing a simple text prompt to genkit.Generate(), pass an array consisting of a media part and a text part. This example specifies an image using a publicly-accessible HTTPS URL.


resp, err := genkit.Generate(ctx, g,
    ai.WithModelName("googleai/gemini-2.0-flash"),
    ai.WithMessages(
        NewUserMessage(
            NewMediaPart("image/jpeg", "https://example.com/photo.jpg"),
            NewTextPart("Compose a poem about this image."),
        ),
    ),
)
You can also pass media data directly by encoding it as a data URL. For example:


image, err := ioutil.ReadFile("photo.jpg")
if err != nil {
    log.Fatal(err)
}

resp, err := genkit.Generate(ctx, g,
    ai.WithModelName("googleai/gemini-2.0-flash"),
    ai.WithMessages(
        NewUserMessage(
            NewMediaPart("image/jpeg", "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(image)),
            NewTextPart("Compose a poem about this image."),
        ),
    ),
)
All models that support media input support both data URLs and HTTPS URLs. Some model plugins add support for other media sources. For example, the Vertex AI plugin also lets you use Cloud Storage (gs://) URLs.

Next steps
Learn more about Genkit
As an app developer, the primary way you influence the output of generative AI models is through prompting. Read Managing prompts with Dotprompt to learn how Genkit helps you develop effective prompts and manage them in your codebase.
Although genkit.Generate() is the nucleus of every generative AI powered application, real-world applications usually require additional work before and after invoking a generative AI model. To reflect this, Genkit introduces the concept of flows, which are defined like functions but add additional features such as observability and simplified deployment. To learn more, see Defining AI workflows.
Advanced LLM use
There are techniques your app can use to reap even more benefit from LLMs.

One way to enhance the capabilities of LLMs is to prompt them with a list of ways they can request more information from you, or request you to perform some action. This is known as tool calling or function calling. Models that are trained to support this capability can respond to a prompt with a specially-formatted response, which indicates to the calling application that it should perform some action and send the result back to the LLM along with the original prompt. Genkit has library functions that automate both the prompt generation and the call-response loop elements of a tool calling implementation. See Tool calling to learn more.
Retrieval-augmented generation (RAG) is a technique used to introduce domain-specific information into a model's output. This is accomplished by inserting relevant information into a prompt before passing it on to the language model. A complete RAG implementation requires you to bring several technologies together: text embedding generation models, vector databases, and large language models. See Retrieval-augmented generation (RAG) to learn how Genkit simplifies the process of coordinating these various elements.

---

Defining AI workflows 

bookmark_border


The core of your app's AI features is generative model requests, but it's rare that you can simply take user input, pass it to the model, and display the model output back to the user. Usually, there are pre- and post-processing steps that must accompany the model call. For example:

Retrieving contextual information to send with the model call.
Retrieving the history of the user's current session, for example in a chat app.
Using one model to reformat the user input in a way that's suitable to pass to another model.
Evaluating the "safety" of a model's output before presenting it to the user.
Combining the output of several models.
Every step of this workflow must work together for any AI-related task to succeed.

In Genkit, you represent this tightly-linked logic using a construction called a flow. Flows are written just like functions, using ordinary Go code, but they add additional capabilities intended to ease the development of AI features:

Type safety: Input and output schemas, which provides both static and runtime type checking.
Integration with developer UI: Debug flows independently of your application code using the developer UI. In the developer UI, you can run flows and view traces for each step of the flow.
Simplified deployment: Deploy flows directly as web API endpoints, using any platform that can host a web app.
Genkit's flows are lightweight and unobtrusive, and don't force your app to conform to any specific abstraction. All of the flow's logic is written in standard Go, and code inside a flow doesn't need to be flow-aware.

Defining and calling flows
In its simplest form, a flow just wraps a function. The following example wraps a function that calls Generate():


menuSuggestionFlow := genkit.DefineFlow(g, "menuSuggestionFlow",
    func(ctx context.Context, theme string) (string, error) {
        resp, err := genkit.Generate(ctx, g,
            ai.WithPrompt("Invent a menu item for a %s themed restaurant.", theme),
        )
        if err != nil {
            return "", err
        }

        return resp.Text(), nil
    })
Just by wrapping your genkit.Generate() calls like this, you add some functionality: Doing so lets you run the flow from the Genkit CLI and from the developer UI, and is a requirement for several of Genkit's features, including deployment and observability (later sections discuss these topics).

Input and output schemas
One of the most important advantages Genkit flows have over directly calling a model API is type safety of both inputs and outputs. When defining flows, you can define schemas, in much the same way as you define the output schema of a genkit.Generate() call; however, unlike with genkit.Generate(), you can also specify an input schema.

Here's a refinement of the last example, which defines a flow that takes a string as input and outputs an object:


type MenuItem struct {
    Name        string `json:"name"`
    Description string `json:"description"`
}

menuSuggestionFlow := genkit.DefineFlow(g, "menuSuggestionFlow",
    func(ctx context.Context, theme string) (MenuItem, error) {
        return genkit.GenerateData[MenuItem](ctx, g,
            ai.WithPrompt("Invent a menu item for a %s themed restaurant.", theme),
        )
    })
Note that the schema of a flow does not necessarily have to line up with the schema of the genkit.Generate() calls within the flow (in fact, a flow might not even contain genkit.Generate() calls). Here's a variation of the example that calls genkit.GenerateData(), but uses the structured output to format a simple string, which the flow returns. Note how we pass MenuItem as a type parameter; this is the equivalent of passing the WithOutputType() option and getting a value of that type in response.


type MenuItem struct {
    Name        string `json:"name"`
    Description string `json:"description"`
}

menuSuggestionMarkdownFlow := genkit.DefineFlow(g, "menuSuggestionMarkdownFlow",
    func(ctx context.Context, theme string) (string, error) {
        item, _, err := genkit.GenerateData[MenuItem](ctx, g,
            ai.WithPrompt("Invent a menu item for a %s themed restaurant.", theme),
        )
        if err != nil {
            return "", err
        }

        return fmt.Sprintf("**%s**: %s", item.Name, item.Description), nil
    })
Calling flows
Once you've defined a flow, you can call it from your Go code:


item, err := menuSuggestionFlow.Run(ctx, "bistro")
The argument to the flow must conform to the input schema.

If you defined an output schema, the flow response will conform to it. For example, if you set the output schema to MenuItem, the flow output will contain its properties:


item, err := menuSuggestionFlow.Run(ctx, "bistro")
if err != nil {
    log.Fatal(err)
}

log.Println(item.DishName)
log.Println(item.Description)
Streaming flows
Flows support streaming using an interface similar to genkit.Generate()'s streaming interface. Streaming is useful when your flow generates a large amount of output, because you can present the output to the user as it's being generated, which improves the perceived responsiveness of your app. As a familiar example, chat-based LLM interfaces often stream their responses to the user as they are generated.

Here's an example of a flow that supports streaming:


type Menu struct {
    Theme  string     `json:"theme"`
    Items  []MenuItem `json:"items"`
}

type MenuItem struct {
    Name        string `json:"name"`
    Description string `json:"description"`
}

menuSuggestionFlow := genkit.DefineStreamingFlow(g, "menuSuggestionFlow",
    func(ctx context.Context, theme string, callback core.StreamCallback[string]) (Menu, error) {
        item, _, err := genkit.GenerateData[MenuItem](ctx, g,
            ai.WithPrompt("Invent a menu item for a %s themed restaurant.", theme),
            ai.WithStreaming(func(ctx context.Context, chunk *ai.ModelResponseChunk) error {
                // Here, you could process the chunk in some way before sending it to
                // the output stream using StreamCallback. In this example, we output
                // the text of the chunk, unmodified.
                return callback(ctx, chunk.Text())
            }),
        )
        if err != nil {
            return nil, err
        }

        return Menu{
            Theme: theme,
            Items: []MenuItem{item},
        }, nil
    })
The string type in StreamCallback[string] specifies the type of values your flow streams. This does not necessarily need to be the same type as the return type, which is the type of the flow's complete output (Menu in this example).

In this example, the values streamed by the flow are directly coupled to the values streamed by the genkit.Generate() call inside the flow. Although this is often the case, it doesn't have to be: you can output values to the stream using the callback as often as is useful for your flow.

Calling streaming flows
Streaming flows can be run like non-streaming flows with menuSuggestionFlow.Run(ctx, "bistro") or they can be streamed:


streamCh, err := menuSuggestionFlow.Stream(ctx, "bistro")
if err != nil {
    log.Fatal(err)
}

for result := range streamCh {
    if result.Err != nil {
        log.Fatal("Stream error: %v", result.Err)
    }
    if result.Done {
        log.Printf("Menu with %s theme:\n", result.Output.Theme)
        for item := range result.Output.Items {
            log.Println(" - %s: %s", item.Name, item.Description)
        }
    } else {
        log.Println("Stream chunk:", result.Stream)
    }
}
Running flows from the command line
You can run flows from the command line using the Genkit CLI tool:


genkit flow:run menuSuggestionFlow '"French"'
For streaming flows, you can print the streaming output to the console by adding the -s flag:


genkit flow:run menuSuggestionFlow '"French"' -s
Running a flow from the command line is useful for testing a flow, or for running flows that perform tasks needed on an ad hoc basis—for example, to run a flow that ingests a document into your vector database.

Debugging flows
One of the advantages of encapsulating AI logic within a flow is that you can test and debug the flow independently from your app using the Genkit developer UI.

The developer UI relies on the Go app continuing to run, even if the logic has completed. If you are just getting started and Genkit is not part of a broader app, add select {} as the last line of main() to prevent the app from shutting down so that you can inspect it in the UI.

To start the developer UI, run the following command from your project directory:


genkit start -- go run .
From the Run tab of developer UI, you can run any of the flows defined in your project:

Screenshot of the Flow runner

After you've run a flow, you can inspect a trace of the flow invocation by either clicking View trace or looking at the Inspect tab.

Deploying flows
You can deploy your flows directly as web API endpoints, ready for you to call from your app clients. Deployment is discussed in detail on several other pages, but this section gives brief overviews of your deployment options.

net/http Server
To deploy a flow using any Go hosting platform, such as Cloud Run, define your flow using DefineFlow() and start a net/http server with the provided flow handler:


import (
    "context"
    "log"
    "net/http"

    "github.com/firebase/genkit/go/genkit"
    "github.com/firebase/genkit/go/plugins/googlegenai"
    "github.com/firebase/genkit/go/plugins/server"
)

func main() {
    ctx := context.Background()

    g, err := genkit.Init(ctx, genkit.WithPlugins(&googlegenai.GoogleAI{}))
    if err != nil {
      log.Fatal(err)
    }

    menuSuggestionFlow := genkit.DefineFlow(g, "menuSuggestionFlow",
        func(ctx context.Context, theme string) (MenuItem, error) {
            // Flow implementation...
        })

    mux := http.NewServeMux()
    mux.HandleFunc("POST /menuSuggestionFlow", genkit.Handler(menuSuggestionFlow))
    log.Fatal(server.Start(ctx, "127.0.0.1:3400", mux))
}
server.Start() is an optional helper function that starts the server and manages its lifecycle, including capturing interrupt signals to ease local development, but you may use your own method.

To serve all the flows defined in your codebase, you can use ListFlows():


mux := http.NewServeMux()
for _, flow := range genkit.ListFlows(g) {
    mux.HandleFunc("POST /"+flow.Name(), genkit.Handler(flow))
}
log.Fatal(server.Start(ctx, "127.0.0.1:3400", mux))
You can call a flow endpoint with a POST request as follows:


curl -X POST "http://localhost:3400/menuSuggestionFlow" \
    -H "Content-Type: application/json" -d '{"data": "banana"}'
Other server frameworks
You can also use other server frameworks to deploy your flows. For example, you can use Gin with just a few lines:


router := gin.Default()
for _, flow := range genkit.ListFlows(g) {
    router.POST("/"+flow.Name(), func(c *gin.Context) {
        genkit.Handler(flow)(c.Writer, c.Request)
    })
}
log.Fatal(router.Run(":3400"))
For information on deploying to specific platforms, see Genkit with Cloud Run.

---

Managing prompts with Dotprompt 

bookmark_border


Prompt engineering is the primary way that you, as an app developer, influence the output of generative AI models. For example, when using LLMs, you can craft prompts that influence the tone, format, length, and other characteristics of the models' responses.

The way you write these prompts will depend on the model you're using; a prompt written for one model might not perform well when used with another model. Similarly, the model parameters you set (temperature, top-k, and so on) will also affect output differently depending on the model.

Getting all three of these factors—the model, the model parameters, and the prompt—working together to produce the output you want is rarely a trivial process and often involves substantial iteration and experimentation. Genkit provides a library and file format called Dotprompt, that aims to make this iteration faster and more convenient.

Dotprompt is designed around the premise that prompts are code. You define your prompts along with the models and model parameters they're intended for separately from your application code. Then, you (or, perhaps someone not even involved with writing application code) can rapidly iterate on the prompts and model parameters using the Genkit Developer UI. Once your prompts are working the way you want, you can import them into your application and run them using Genkit.

Your prompt definitions each go in a file with a .prompt extension. Here's an example of what these files look like:


---
model: googleai/gemini-1.5-flash
config:
  temperature: 0.9
input:
  schema:
    location: string
    style?: string
    name?: string
  default:
    location: a restaurant
---

You are the world's most welcoming AI assistant and are currently working at {{location}}.

Greet a guest{{#if name}} named {{name}}{{/if}}{{#if style}} in the style of {{style}}{{/if}}.
The portion in the triple-dashes is YAML front matter, similar to the front matter format used by GitHub Markdown and Jekyll; the rest of the file is the prompt, which can optionally use Handlebars templates. The following sections will go into more detail about each of the parts that make a .prompt file and how to use them.

Before you begin
Before reading this page, you should be familiar with the content covered on the Generating content with AI models page.

If you want to run the code examples on this page, first complete the steps in the Get started guide. All of the examples assume that you have already installed Genkit as a dependency in your project.

Creating prompt files
Although Dotprompt provides several different ways to create and load prompts, it's optimized for projects that organize their prompts as .prompt files within a single directory (or subdirectories thereof). This section shows you how to create and load prompts using this recommended setup.

Creating a prompt directory
The Dotprompt library expects to find your prompts in a directory at your project root and automatically loads any prompts it finds there. By default, this directory is named prompts. For example, using the default directory name, your project structure might look something like this:


your-project/
├── prompts/
│   └── hello.prompt
├── main.go
├── go.mod
└── go.sum
If you want to use a different directory, you can specify it when you configure Genkit:


g, err := genkit.Init(ctx.Background(), ai.WithPromptDir("./llm_prompts"))
Creating a prompt file
There are two ways to create a .prompt file: using a text editor, or with the developer UI.

Using a text editor
If you want to create a prompt file using a text editor, create a text file with the .prompt extension in your prompts directory: for example, prompts/hello.prompt.

Here is a minimal example of a prompt file:


---
model: vertexai/gemini-1.5-flash
---
You are the world's most welcoming AI assistant. Greet the user and offer your
assistance.
The portion in the dashes is YAML front matter, similar to the front matter format used by GitHub Markdown and Jekyll; the rest of the file is the prompt, which can optionally use Handlebars templates. The front matter section is optional, but most prompt files will at least contain metadata specifying a model. The remainder of this page shows you how to go beyond this, and make use of Dotprompt's features in your prompt files.

Using the developer UI
You can also create a prompt file using the model runner in the developer UI. Start with application code that imports the Genkit library and configures it to use the model plugin you're interested in. For example:


import (
    "context"

    "github.com/firebase/genkit/go/ai"
    "github.com/firebase/genkit/go/genkit"
    "github.com/firebase/genkit/go/plugins/googlegenai"
)

func main() {
    g, err := genkit.Init(context.Background(), ai.WithPlugins(&googlegenai.GoogleAI{}))
    if err != nil {
        log.Fatal(err)
    }

    // Blocks end of program execution to use the developer UI.
    select {}
}
Load the developer UI in the same project:


genkit start -- go run .
In the Model section, choose the model you want to use from the list of models provided by the plugin.

Genkit developer UI model runner

Then, experiment with the prompt and configuration until you get results you're happy with. When you're ready, press the Export button and save the file to your prompts directory.

Running prompts
After you've created prompt files, you can run them from your application code, or using the tooling provided by Genkit. Regardless of how you want to run your prompts, first start with application code that imports the Genkit library and the model plugins you're interested in. For example:


import (
    "context"

      "github.com/firebase/genkit/go/ai"
    "github.com/firebase/genkit/go/genkit"
    "github.com/firebase/genkit/go/plugins/googlegenai"
)

func main() {
    g, err := genkit.Init(context.Background(), ai.WithPlugins(&googlegenai.GoogleAI{}))
    if err != nil {
        log.Fatal(err)
    }

    // Blocks end of program execution to use the developer UI.
    select {}
}
If you're storing your prompts in a directory other than the default, be sure to specify it when you configure Genkit.

Run prompts from code
To use a prompt, first load it using the genkit.LookupPrompt() function:


helloPrompt := genkit.LookupPrompt(g, "hello")
An executable prompt has similar options to that of genkit.Generate() and many of them are overridable at execution time, including things like input (see the section about specifying input schemas), configuration, and more:


resp, err := helloPrompt.Execute(context.Background(),
    ai.WithModelName("googleai/gemini-2.0-flash"),
    ai.WithInput(map[string]any{"name": "John"}),
    ai.WithConfig(&googlegenai.GeminiConfig{Temperature: 0.5})
)
Any parameters you pass to the prompt call will override the same parameters specified in the prompt file.

See Generate content with AI models for descriptions of the available options.

Using the developer UI
As you're refining your app's prompts, you can run them in the Genkit developer UI to quickly iterate on prompts and model configurations, independently from your application code.

Load the developer UI from your project directory:


genkit start -- go run .
Genkit developer UI prompt runner

Once you've loaded prompts into the developer UI, you can run them with different input values, and experiment with how changes to the prompt wording or the configuration parameters affect the model output. When you're happy with the result, you can click the Export prompt button to save the modified prompt back into your project directory.

Model configuration
In the front matter block of your prompt files, you can optionally specify model configuration values for your prompt:


---
model: googleai/gemini-2.0-flash
config:
  temperature: 1.4
  topK: 50
  topP: 0.4
  maxOutputTokens: 400
  stopSequences:
    -   "<end>"
    -   "<fin>"
---
These values map directly to the WithConfig() option accepted by the executable prompt:


resp, err := helloPrompt.Execute(context.Background(),
    ai.WithConfig(&googlegenai.GeminiConfig{
        Temperature:     1.4,
        TopK:            50,
        TopP:            0.4,
        MaxOutputTokens: 400,
        StopSequences:   []string{"<end>", "<fin>"},
    }))
See Generate content with AI models for descriptions of the available options.

Input and output schemas
You can specify input and output schemas for your prompt by defining them in the front matter section. These schemas are used in much the same way as those passed to a genkit.Generate() request or a flow definition:


---
model: googleai/gemini-2.0-flash
input:
  schema:
    theme?: string
  default:
    theme: "pirate"
output:
  schema:
    dishname: string
    description: string
    calories: integer
    allergens(array): string
---
Invent a menu item for a {{theme}} themed
restaurant.
This code produces the following structured output:


menuPrompt = genkit.LookupPrompt(g, "menu")
if menuPrompt == nil {
    log.Fatal("no prompt named 'menu' found")
}

resp, err := menuPrompt.Execute(context.Background(),
    ai.WithInput(map[string]any{"theme": "medieval"}),
)
if err != nil {
    log.Fatal(err)
}

var output map[string]any
if err := resp.Output(&output); err != nil {
    log.Fatal(err)
}

log.Println(output["dishname"])
log.Println(output["description"])
You have several options for defining schemas in a .prompt file: Dotprompt's own schema definition format, Picoschema; standard JSON Schema; or, as references to schemas defined in your application code. The following sections describe each of these options in more detail.

Picoschema
The schemas in the example above are defined in a format called Picoschema. Picoschema is a compact, YAML-optimized schema definition format that simplifies defining the most important attributes of a schema for LLM usage. Here's a longer example of a schema, which specifies the information an app might store about an article:


schema:
  title: string # string, number, and boolean types are defined like this
  subtitle?: string # optional fields are marked with a `?`
  draft?: boolean, true when in draft state
  status?(enum, approval status): [PENDING, APPROVED]
  date: string, the date of publication e.g. '2024-04-09' # descriptions follow a comma
  tags(array, relevant tags for article): string # arrays are denoted via parentheses
  authors(array):
    name: string
    email?: string
  metadata?(object): # objects are also denoted via parentheses
    updatedAt?: string, ISO timestamp of last update
    approvedBy?: integer, id of approver
  extra?: any, arbitrary extra data
  (*): string, wildcard field
The above schema is equivalent to the following Go type:


type Article struct {
    Title    string   `json:"title"`
    Subtitle string   `json:"subtitle,omitempty" jsonschema:"required=false"`
    Draft    bool     `json:"draft,omitempty"`  // True when in draft state
    Status   string   `json:"status,omitempty" jsonschema:"enum=PENDING,enum=APPROVED"` // Approval status
    Date     string   `json:"date"`   // The date of publication e.g. '2025-04-07'
    Tags     []string `json:"tags"`   // Relevant tags for article
    Authors  []struct {
      Name  string `json:"name"`
      Email string `json:"email,omitempty"`
    } `json:"authors"`
    Metadata struct {
      UpdatedAt  string `json:"updatedAt,omitempty"`  // ISO timestamp of last update
      ApprovedBy int    `json:"approvedBy,omitempty"` // ID of approver
    } `json:"metadata,omitempty"`
    Extra any `json:"extra"` // Arbitrary extra data
}
Picoschema supports scalar types string, integer, number, boolean, and any. Objects, arrays, and enums are denoted by a parenthetical after the field name.

Objects defined by Picoschema have all properties required unless denoted optional by ?, and don't allow additional properties. When a property is marked as optional, it is also made nullable to provide more leniency for LLMs to return null instead of omitting a field.

In an object definition, the special key (*) can be used to declare a "wildcard" field definition. This will match any additional properties not supplied by an explicit key.

JSON Schema
Picoschema does not support many of the capabilities of full JSON schema. If you require more robust schemas, you may supply a JSON Schema instead:


output:
  schema:
    type: object
    properties:
      field1:
        type: number
        minimum: 20
Prompt templates
The portion of a .prompt file that follows the front matter (if present) is the prompt itself, which will be passed to the model. While this prompt could be a simple text string, very often you will want to incorporate user input into the prompt. To do so, you can specify your prompt using the Handlebars templating language. Prompt templates can include placeholders that refer to the values defined by your prompt's input schema.

You already saw this in action in the section on input and output schemas:


---
model: googleai/gemini-2.0-flash
input:
  schema:
    theme?: string
  default:
    theme: "pirate"
output:
  schema:
    dishname: string
    description: string
    calories: integer
    allergens(array): string
---
Invent a menu item for a {{theme}} themed restaurant.
In this example, the Handlebars expression, {{theme}}, resolves to the value of the input's theme property when you run the prompt. To pass input to the prompt, call the prompt as in the following example:


menuPrompt = genkit.LookupPrompt(g, "menu")

resp, err := menuPrompt.Execute(context.Background(),
    ai.WithInput(map[string]any{"theme": "medieval"}),
)
Note that because the input schema declared the theme property to be optional and provided a default, you could have omitted the property, and the prompt would have resolved using the default value.

Handlebars templates also support some limited logical constructs. For example, as an alternative to providing a default, you could define the prompt using Handlebars's #if helper:


---
model: googleai/gemini-2.0-flash
input:
  schema:
    theme?: string
---
Invent a menu item for a {{#if theme}}{{theme}}{else}themed{{/else}} restaurant.
In this example, the prompt renders as "Invent a menu item for a restaurant" when the theme property is unspecified.

See the Handlebars documentation for information on all of the built-in logical helpers.

In addition to properties defined by your input schema, your templates can also refer to values automatically defined by Genkit. The next few sections describe these automatically-defined values and how you can use them.

Multi-message prompts
By default, Dotprompt constructs a single message with a "user" role. However, some prompts, such as a system prompt, are best expressed as combinations of multiple messages.

The {{role}} helper provides a straightforward way to construct multi-message prompts:


---
model: vertexai/gemini-2.0-flash
input:
  schema:
    userQuestion: string
---
{{role "system"}}
You are a helpful AI assistant that really loves to talk about food. Try to work
food items into all of your conversations.

{{role "user"}}
{{userQuestion}}
Multi-modal prompts
For models that support multimodal input, such as images alongside text, you can use the {{media}} helper:


---
model: vertexai/gemini-2.0-flash
input:
  schema:
    photoUrl: string
---
Describe this image in a detailed paragraph:

{{media url=photoUrl}}
The URL can be https: or base64-encoded data: URIs for "inline" image usage. In code, this would be:


multimodalPrompt = genkit.LookupPrompt(g, "multimodal")

resp, err := multimodalPrompt.Execute(context.Background(),
    ai.WithInput(map[string]any{"photoUrl": "https://example.com/photo.jpg"}),
)
See also Multimodal input, on the Generating content with AI models page, for an example of constructing a data: URL.

Partials
Partials are reusable templates that can be included inside any prompt. Partials can be especially helpful for related prompts that share common behavior.

When loading a prompt directory, any file prefixed with an underscore (_) is considered a partial. So a file _personality.prompt might contain:


You should speak like a {{#if style}}{{style}}{else}helpful assistant.{{/else}}.
This can then be included in other prompts:


---
model: googleai/gemini-2.0-flash
input:
  schema:
    name: string
    style?: string
---
{{ role "system" }}
{{>personality style=style}}

{{ role "user" }}
Give the user a friendly greeting.

User's Name: {{name}}
Partials are inserted using the {{>NAME_OF_PARTIAL args...}} syntax. If no arguments are provided to the partial, it executes with the same context as the parent prompt.

Partials accept named arguments or a single positional argument representing the context. This can be helpful for tasks such as rendering members of a list.

_destination.prompt


-   {{name}} ({{country}})
chooseDestination.prompt


---
model: googleai/gemini-2.0-flash
input:
  schema:
    destinations(array):
      name: string
      country: string
---
Help the user decide between these vacation destinations:

{{#each destinations}}
{{>destination this}}
{{/each}}
Defining partials in code
You can also define partials in code using genkit.DefinePartial():


genkit.DefinePartial(g, "personality", "Talk like a {{#if style}}{{style}}{{else}}helpful assistant{{/if}}.")
Code-defined partials are available in all prompts.

Defining Custom Helpers
You can define custom helpers to process and manage data inside of a prompt. Helpers are registered globally using genkit.DefineHelper():


genkit.DefineHelper(g, "shout", func(input string) string {
    return strings.ToUpper(input)
})
Once a helper is defined you can use it in any prompt:


---
model: googleai/gemini-2.0-flash
input:
  schema:
    name: string
---

HELLO, {{shout name}}!!!
Prompt variants
Because prompt files are just text, you can (and should!) commit them to your version control system, simplifying the process of comparing changes over time. Often, tweaked versions of prompts can only be fully tested in a production environment side-by-side with existing versions. Dotprompt supports this through its variants feature.

To create a variant, create a [name].[variant].prompt file. For example, if you were using Gemini 2.0 Flash in your prompt but wanted to see if Gemini 2.5 Pro would perform better, you might create two files:

myPrompt.prompt: the "baseline" prompt
myPrompt.gemini25pro.prompt: a variant named gemini25pro
To use a prompt variant, specify the variant option when loading:


myPrompt := genkit.LookupPrompt(g, "myPrompt.gemini25Pro")
The name of the variant is included in the metadata of generation traces, so you can compare and contrast actual performance between variants in the Genkit trace inspector.

Defining prompts in code
All of the examples discussed so far have assumed that your prompts are defined in individual .prompt files in a single directory (or subdirectories thereof), accessible to your app at runtime. Dotprompt is designed around this setup, and its authors consider it to be the best developer experience overall.

However, if you have use cases that are not well supported by this setup, you can also define prompts in code using the genkit.DefinePrompt() function:


type GeoQuery struct {
    CountryCount int `json:"countryCount"`
}

type CountryList struct {
    Countries []string `json:"countries"`
}

geographyPrompt, err := genkit.DefinePrompt(
    g, "GeographyPrompt",
    ai.WithSystem("You are a geography teacher. Respond only when the user asks about geography."),
    ai.WithPrompt("Give me the {{countryCount}} biggest countries in the world by inhabitants."),
    ai.WithConfig(&googlegenai.GeminiConfig{Temperature: 0.5}),
    ai.WithInputType(GeoQuery{CountryCount: 10}) // Defaults to 10.
    ai.WithOutputType(CountryList{}),
)
if err != nil {
    log.Fatal(err)
}

resp, err := geographyPrompt.Execute(context.Background(), ai.WithInput(GeoQuery{CountryCount: 15}))
if err != nil {
    log.Fatal(err)
}

var list CountryList
if err := resp.Output(&list); err != nil {
    log.Fatal(err)
}

log.Println("Countries: %s", list.Countries)
Prompts may also be rendered into a GenerateActionOptions which may then be processed and passed into genkit.GenerateWithRequest():


actionOpts, err := geographyPrompt.Render(ctx, ai.WithInput(GeoQuery{CountryCount: 15}))
if err != nil {
    log.Fatal(err)
}

// Do something with the value...
actionOpts.Config = &googlegenai.GeminiConfig{Temperature: 0.8}

resp, err := genkit.GenerateWithRequest(ctx, g, actionOpts, nil, nil) // No middleware or streaming
Note that all prompt options carry over to GenerateActionOptions with the exception of WithMiddleware(), which must be passed separately if using Prompt.Render() instead of Prompt.Execute().

---

Tool calling 

bookmark_border
Tool calling, also known as function calling, is a structured way to give LLMs the ability to make requests back to the application that called it. You define the tools you want to make available to the model, and the model will make tool requests to your app as necessary to fulfill the prompts you give it.

The use cases of tool calling generally fall into a few themes:

Giving an LLM access to information it wasn't trained with

Frequently changing information, such as a stock price or the current weather.
Information specific to your app domain, such as product information or user profiles.
Note the overlap with retrieval augmented generation (RAG), which is also a way to let an LLM integrate factual information into its generations. RAG is a heavier solution that is most suited when you have a large amount of information or the information that's most relevant to a prompt is ambiguous. On the other hand, if a function call or database lookup is all that's necessary for retrieving the information the LLM needs, tool calling is more appropriate.

Introducing a degree of determinism into an LLM workflow

Performing calculations that the LLM cannot reliably complete itself.
Forcing an LLM to generate verbatim text under certain circumstances, such as when responding to a question about an app's terms of service.
Performing an action when initiated by an LLM

Turning on and off lights in an LLM-powered home assistant
Reserving table reservations in an LLM-powered restaurant agent
Before you begin
If you want to run the code examples on this page, first complete the steps in the Get started guide. All of the examples assume that you have already set up a project with Genkit dependencies installed.

This page discusses one of the advanced features of Genkit model abstraction, so before you dive too deeply, you should be familiar with the content on the Generating content with AI models page. You should also be familiar with Genkit's system for defining input and output schemas, which is discussed on the Flows page.

Overview of tool calling
At a high level, this is what a typical tool-calling interaction with an LLM looks like:

The calling application prompts the LLM with a request and also includes in the prompt a list of tools the LLM can use to generate a response.
The LLM either generates a complete response or generates a tool call request in a specific format.
If the caller receives a complete response, the request is fulfilled and the interaction ends; but if the caller receives a tool call, it performs whatever logic is appropriate and sends a new request to the LLM containing the original prompt or some variation of it as well as the result of the tool call.
The LLM handles the new prompt as in Step 2.
For this to work, several requirements must be met:

The model must be trained to make tool requests when it's needed to complete a prompt. Most of the larger models provided through web APIs such as Gemini can do this, but smaller and more specialized models often cannot. Genkit will throw an error if you try to provide tools to a model that doesn't support it.
The calling application must provide tool definitions to the model in the format it expects.
The calling application must prompt the model to generate tool calling requests in the format the application expects.
Tool calling with Genkit
Genkit provides a single interface for tool calling with models that support it. Each model plugin ensures that the last two criteria mentioned in the previous section are met, and the genkit.Generate() function automatically carries out the tool-calling loop described earlier.

Model support
Tool calling support depends on the model, the model API, and the Genkit plugin. Consult the relevant documentation to determine if tool calling is likely to be supported. In addition:

Genkit will throw an error if you try to provide tools to a model that doesn't support it.
If the plugin exports model references, the ModelInfo.Supports.Tools property will indicate if it supports tool calling.
Defining tools
Use the genkit.DefineTool() function to write tool definitions:


import (
    "context"
    "log"

    "github.com/firebase/genkit/go/ai"
    "github.com/firebase/genkit/go/genkit"
    "github.com/firebase/genkit/go/plugins/googlegenai"
)

func main() {
    ctx := context.Background()

    g, err := genkit.Init(ctx,
        genkit.WithPlugins(&googlegenai.GoogleAI{}),
        genkit.WithDefaultModel("googleai/gemini-2.0-flash"),
    )
    if err != nil {
      log.Fatal(err)
    }

    getWeatherTool := genkit.DefineTool(
        g, "getWeather", "Gets the current weather in a given location",
        func(ctx *ai.ToolContext, input struct{
            Location string `jsonschema_description:"Location to get weather for"`
        }) (string, error) {
            // Here, we would typically make an API call or database query. For this
            // example, we just return a fixed value.
            return fmt.Sprintf("The current weather in %s is 63°F and sunny.", input.Location);
        })
}
The syntax here looks just like the genkit.DefineFlow() syntax; however, you must write a description. Take special care with the wording and descriptiveness of the description as it is vital for the LLM to decide to use it appropriately.

Using tools
Include defined tools in your prompts to generate content.

Generate
DefinePrompt
Prompt file

resp, err := genkit.Generate(ctx, g,
    ai.WithPrompt("What is the weather in San Francisco?"),
    ai.WithTools(getWeatherTool),
)
Genkit will automatically handle the tool call if the LLM needs to use the getWeather tool to answer the prompt.

Explicitly handling tool calls
If you want full control over this tool-calling loop, for example to apply more complicated logic, set the WithReturnToolRequests() option to true. Now it's your responsibility to ensure all of the tool requests are fulfilled:


getWeatherTool := genkit.DefineTool(
    g, "getWeather", "Gets the current weather in a given location",
    func(ctx *ai.ToolContext, location struct {
        Location string `jsonschema_description:"Location to get weather for"`
    }) (string, error) {
        // Tool implementation...
        return "sunny", nil
    },
)

resp, err := genkit.Generate(ctx, g,
    ai.WithPrompt("What is the weather in San Francisco?"),
    ai.WithTools(getWeatherTool),
    ai.WithReturnToolRequests(true),
)
if err != nil {
    log.Fatal(err)
}

parts := []*ai.Part{}
for _, req := range resp.ToolRequests() {
    tool := genkit.LookupTool(g, req.Name)
    if tool == nil {
        log.Fatalf("tool %q not found", req.Name)
    }

    output, err := tool.RunRaw(ctx, req.Input)
    if err != nil {
        log.Fatalf("tool %q execution failed: %v", tool.Name(), err)
    }

    parts = append(parts,
        ai.NewToolResponsePart(&ai.ToolResponse{
            Name:   req.Name,
            Ref:    req.Ref,
            Output: output,
        }))
}

resp, err = genkit.Generate(ctx, g,
    ai.WithMessages(append(resp.History(), ai.NewMessage(ai.RoleTool, nil, parts...))...),
)
if err != nil {
    log.Fatal(err)
}

---

Retrieval-augmented generation (RAG) 

bookmark_border


Genkit provides abstractions that help you build retrieval-augmented generation (RAG) flows, as well as plugins that provide integrations with related tools.

What is RAG?
Retrieval-augmented generation is a technique used to incorporate external sources of information into an LLM’s responses. It's important to be able to do so because, while LLMs are typically trained on a broad body of material, practical use of LLMs often requires specific domain knowledge (for example, you might want to use an LLM to answer customers' questions about your company’s products).

One solution is to fine-tune the model using more specific data. However, this can be expensive both in terms of compute cost and in terms of the effort needed to prepare adequate training data.

In contrast, RAG works by incorporating external data sources into a prompt at the time it's passed to the model. For example, you could imagine the prompt, "What is Bart's relationship to Lisa?" might be expanded ("augmented") by prepending some relevant information, resulting in the prompt, "Homer and Marge's children are named Bart, Lisa, and Maggie. What is Bart's relationship to Lisa?"

This approach has several advantages:

It can be more cost effective because you don't have to retrain the model.
You can continuously update your data source and the LLM can immediately make use of the updated information.
You now have the potential to cite references in your LLM's responses.
On the other hand, using RAG naturally means longer prompts, and some LLM API services charge for each input token you send. Ultimately, you must evaluate the cost tradeoffs for your applications.

RAG is a very broad area and there are many different techniques used to achieve the best quality RAG. The core Genkit framework offers two main abstractions to help you do RAG:

Indexers: add documents to an "index".
Embedders: transforms documents into a vector representation
Retrievers: retrieve documents from an "index", given a query.
These definitions are broad on purpose because Genkit is un-opinionated about what an "index" is or how exactly documents are retrieved from it. Genkit only provides a Document format and everything else is defined by the retriever or indexer implementation provider.

Indexers
The index is responsible for keeping track of your documents in such a way that you can quickly retrieve relevant documents given a specific query. This is most often accomplished using a vector database, which indexes your documents using multidimensional vectors called embeddings. A text embedding (opaquely) represents the concepts expressed by a passage of text; these are generated using special-purpose ML models. By indexing text using its embedding, a vector database is able to cluster conceptually related text and retrieve documents related to a novel string of text (the query).

Before you can retrieve documents for the purpose of generation, you need to ingest them into your document index. A typical ingestion flow does the following:

Split up large documents into smaller documents so that only relevant portions are used to augment your prompts – "chunking". This is necessary because many LLMs have a limited context window, making it impractical to include entire documents with a prompt.

Genkit doesn't provide built-in chunking libraries; however, there are open source libraries available that are compatible with Genkit.

Generate embeddings for each chunk. Depending on the database you're using, you might explicitly do this with an embedding generation model, or you might use the embedding generator provided by the database.

Add the text chunk and its index to the database.

You might run your ingestion flow infrequently or only once if you are working with a stable source of data. On the other hand, if you are working with data that frequently changes, you might continuously run the ingestion flow (for example, in a Cloud Firestore trigger, whenever a document is updated).

Embedders
An embedder is a function that takes content (text, images, audio, etc.) and creates a numeric vector that encodes the semantic meaning of the original content. As mentioned above, embedders are leveraged as part of the process of indexing. However, they can also be used independently to create embeddings without an index.

Retrievers
A retriever is a concept that encapsulates logic related to any kind of document retrieval. The most popular retrieval cases typically include retrieval from vector stores. However, in Genkit a retriever can be any function that returns data.

To create a retriever, you can use one of the provided implementations or create your own.

Supported indexers, retrievers, and embedders
Genkit provides indexer and retriever support through its plugin system. The following plugins are officially supported:

Pinecone cloud vector database
In addition, Genkit supports the following vector stores through predefined code templates, which you can customize for your database configuration and schema:

PostgreSQL with pgvector
Embedding model support is provided through the following plugins:

Plugin	Models
Google Generative AI	Text embedding
Defining a RAG Flow
The following examples show how you could ingest a collection of restaurant menu PDF documents into a vector database and retrieve them for use in a flow that determines what food items are available.

Install dependencies
In this example, we will use the textsplitter library from langchaingo and the ledongthuc/pdf PDF parsing Library:


go get github.com/tmc/langchaingo/textsplitter
go get github.com/ledongthuc/pdf
Define an Indexer
The following example shows how to create an indexer to ingest a collection of PDF documents and store them in a local vector database.

It uses the local file-based vector similarity retriever that Genkit provides out-of-the box for simple testing and prototyping. Do not use this in production.

Create the indexer

// Import Genkit's file-based vector retriever, (Don't use in production.)
import "github.com/firebase/genkit/go/plugins/localvec"

// Vertex AI provides the text-embedding-004 embedder model.
import "github.com/firebase/genkit/go/plugins/vertexai"

ctx := context.Background()

g, err := genkit.Init(ctx, genkit.WithPlugins(&googlegenai.VertexAI{}))
if err != nil {
    log.Fatal(err)
}

if err = localvec.Init(); err != nil {
    log.Fatal(err)
}

menuPDFIndexer, _, err := localvec.DefineIndexerAndRetriever(g, "menuQA",
      localvec.Config{Embedder: googlegenai.VertexAIEmbedder(g, "text-embedding-004")})
if err != nil {
    log.Fatal(err)
}
Create chunking config
This example uses the textsplitter library which provides a simple text splitter to break up documents into segments that can be vectorized.

The following definition configures the chunking function to return document segments of 200 characters, with an overlap between chunks of 20 characters.


splitter := textsplitter.NewRecursiveCharacter(
    textsplitter.WithChunkSize(200),
    textsplitter.WithChunkOverlap(20),
)
More chunking options for this library can be found in the langchaingo documentation.

Define your indexer flow

genkit.DefineFlow(
    g, "indexMenu",
    func(ctx context.Context, path string) (any, error) {
        // Extract plain text from the PDF. Wrap the logic in Run so it
        // appears as a step in your traces.
        pdfText, err := genkit.Run(ctx, "extract", func() (string, error) {
            return readPDF(path)
        })
        if err != nil {
            return nil, err
        }

        // Split the text into chunks. Wrap the logic in Run so it appears as a
        // step in your traces.
        docs, err := genkit.Run(ctx, "chunk", func() ([]*ai.Document, error) {
            chunks, err := splitter.SplitText(pdfText)
            if err != nil {
                return nil, err
            }

            var docs []*ai.Document
            for _, chunk := range chunks {
                docs = append(docs, ai.DocumentFromText(chunk, nil))
            }
            return docs, nil
        })
        if err != nil {
            return nil, err
        }

        // Add chunks to the index.
        err = ai.Index(ctx, menuPDFIndexer, ai.WithDocs(docs...))
        return nil, err
    },
)

// Helper function to extract plain text from a PDF. Excerpted from
// https://github.com/ledongthuc/pdf
func readPDF(path string) (string, error) {
    f, r, err := pdf.Open(path)
    if f != nil {
        defer f.Close()
    }
    if err != nil {
        return "", err
    }

    reader, err := r.GetPlainText()
    if err != nil {
        return "", err
    }

    bytes, err := io.ReadAll(reader)
    if err != nil {
        return "", err
    }

    return string(bytes), nil
}
Run the indexer flow

genkit flow:run indexMenu "'menu.pdf'"
After running the indexMenu flow, the vector database will be seeded with documents and ready to be used in Genkit flows with retrieval steps.

Define a flow with retrieval
The following example shows how you might use a retriever in a RAG flow. Like the indexer example, this example uses Genkit's file-based vector retriever, which you should not use in production.


ctx := context.Background()

g, err := genkit.Init(ctx, genkit.WithPlugins(&googlegenai.VertexAI{}))
if err != nil {
    log.Fatal(err)
}

if err = localvec.Init(); err != nil {
    log.Fatal(err)
}

model := googlegenai.VertexAIModel(g, "gemini-1.5-flash")

_, menuPdfRetriever, err := localvec.DefineIndexerAndRetriever(
    g, "menuQA", localvec.Config{Embedder: googlegenai.VertexAIEmbedder(g, "text-embedding-004")},
)
if err != nil {
    log.Fatal(err)
}

genkit.DefineFlow(
  g, "menuQA",
  func(ctx context.Context, question string) (string, error) {
    // Retrieve text relevant to the user's question.
    resp, err := ai.Retrieve(ctx, menuPdfRetriever, ai.WithTextDocs(question))


    if err != nil {
        return "", err
    }

    // Call Generate, including the menu information in your prompt.
    return genkit.GenerateText(ctx, g,
        ai.WithModelName("googleai/gemini-2.0-flash"),
        ai.WithDocs(resp.Documents),
        ai.WithSystem(`
You are acting as a helpful AI assistant that can answer questions about the
food available on the menu at Genkit Grub Pub.
Use only the context provided to answer the question. If you don't know, do not
make up an answer. Do not add or change items on the menu.`)
        ai.WithPrompt(question),
  })
Write your own indexers and retrievers
It's also possible to create your own retriever. This is useful if your documents are managed in a document store that is not supported in Genkit (eg: MySQL, Google Drive, etc.). The Genkit SDK provides flexible methods that let you provide custom code for fetching documents.

You can also define custom retrievers that build on top of existing retrievers in Genkit and apply advanced RAG techniques (such as reranking or prompt extension) on top.

For example, suppose you have a custom re-ranking function you want to use. The following example defines a custom retriever that applies your function to the menu retriever defined earlier:


type CustomMenuRetrieverOptions struct {
    K          int
    PreRerankK int
}

advancedMenuRetriever := genkit.DefineRetriever(
    g, "custom", "advancedMenuRetriever",
    func(ctx context.Context, req *ai.RetrieverRequest) (*ai.RetrieverResponse, error) {
        // Handle options passed using our custom type.
        opts, _ := req.Options.(CustomMenuRetrieverOptions)
        // Set fields to default values when either the field was undefined
        // or when req.Options is not a CustomMenuRetrieverOptions.
        if opts.K == 0 {
            opts.K = 3
        }
        if opts.PreRerankK == 0 {
            opts.PreRerankK = 10
        }

        // Call the retriever as in the simple case.
        resp, err := ai.Retrieve(ctx, menuPDFRetriever,
            ai.WithDocs(req.Query),
            ai.WithConfig(ocalvec.RetrieverOptions{K: opts.PreRerankK}),
        )
        if err != nil {
            return nil, err
        }

        // Re-rank the returned documents using your custom function.
        rerankedDocs := rerank(response.Documents)
        response.Documents = rerankedDocs[:opts.K]

        return response, nil
    },
)

---

Evaluation 

bookmark_border


Evaluation is a form of testing that helps you validate your LLM's responses and ensure they meet your quality bar.

Genkit supports third-party evaluation tools through plugins, paired with powerful observability features that provide insight into the runtime state of your LLM-powered applications. Genkit tooling helps you automatically extract data including inputs, outputs, and information from intermediate steps to evaluate the end-to-end quality of LLM responses as well as understand the performance of your system's building blocks.

Types of evaluation
Genkit supports two types of evaluation:

Inference-based evaluation: This type of evaluation runs against a collection of pre-determined inputs, assessing the corresponding outputs for quality.

This is the most common evaluation type, suitable for most use cases. This approach tests a system's actual output for each evaluation run.

You can perform the quality assessment manually, by visually inspecting the results. Alternatively, you can automate the assessment by using an evaluation metric.

Raw evaluation: This type of evaluation directly assesses the quality of inputs without any inference. This approach typically is used with automated evaluation using metrics. All required fields for evaluation (e.g., input, context, output and reference) must be present in the input dataset. This is useful when you have data coming from an external source (e.g., collected from your production traces) and you want to have an objective measurement of the quality of the collected data.

For more information, see the Advanced use section of this page.

This section explains how to perform inference-based evaluation using Genkit.

Quick start
Perform these steps to get started quickly with Genkit.

Setup
Use an existing Genkit app or create a new one by following our Get started guide.
Add the following code to define a simple RAG application to evaluate. For this guide, we use a dummy retriever that always returns the same documents.

import (
    "context"
    "fmt"
    "log"

    "github.com/firebase/genkit/go/ai"
    "github.com/firebase/genkit/go/genkit"
    "github.com/firebase/genkit/go/plugins/googlegenai"
)

func main() {
    ctx := context.Background()

    // Initialize Genkit
    g, err := genkit.Init(ctx,
        genkit.WithPlugins(&googlegenai.GoogleAI{}),
        genkit.WithDefaultModel("googleai/gemini-2.0-flash"),
    )
    if err != nil {
        log.Fatalf("Genkit initialization error: %v", err)
    }

    // Dummy retriever that always returns the same facts
    dummyRetrieverFunc := func(ctx context.Context, req *ai.RetrieverRequest) (*ai.RetrieverResponse, error) {
        facts := []string{
            "Dog is man's best friend",
            "Dogs have evolved and were domesticated from wolves",
        }
        // Just return facts as documents.
        var docs []*ai.Document
        for _, fact := range facts {
            docs = append(docs, ai.DocumentFromText(fact, nil))
        }
        return &ai.RetrieverResponse{Documents: docs}, nil
    }
    factsRetriever := genkit.DefineRetriever(g, "local", "dogFacts", dummyRetrieverFunc)

    m := googlegenai.GoogleAIModel(g, "gemini-2.0-flash")
    if m == nil {
        log.Fatal("failed to find model")
    }

    // A simple question-answering flow
    genkit.DefineFlow(g, "qaFlow", func(ctx context.Context, query string) (string, error) {
        factDocs, err := ai.Retrieve(ctx, factsRetriever, ai.WithTextDocs(query))
        if err != nil {
            return "", fmt.Errorf("retrieval failed: %w", err)
        }
        llmResponse, err := genkit.Generate(ctx, g,
            ai.WithModelName("googleai/gemini-2.0-flash"),
            ai.WithPrompt("Answer this question with the given context: %s", query),
            ai.WithDocs(factDocs.Documents...)
        )
        if err != nil {
            return "", fmt.Errorf("generation failed: %w", err)
        }
        return llmResponse.Text(), nil
    })
}
You can optionally add evaluation metrics to your application to use while evaluating. This guide uses the EvaluatorRegex metric from the evaluators package.

import (
    "github.com/firebase/genkit/go/plugins/evaluators"
)

func main() {
  // ...

  metrics := []evaluators.MetricConfig{
      {
        MetricType: evaluators.EvaluatorRegex,
      },
  }

  // Initialize Genkit
  g, err := genkit.Init(ctx,
      genkit.WithPlugins(
          &googlegenai.GoogleAI{},
          &evaluators.GenkitEval{Metrics: metrics}        // Add this plugin
      ),
      genkit.WithDefaultModel("googleai/gemini-2.0-flash"),
  )
}
Note: Ensure that the evaluators package is installed in your go project:

go get github.com/firebase/genkit/go/plugins/evaluators
Start your Genkit application.

genkit start -- go run main.go
Create a dataset
Create a dataset to define the examples we want to use for evaluating our flow.

Go to the Dev UI at http://localhost:4000 and click the Datasets button to open the Datasets page.

Click the Create Dataset button to open the create dataset dialog.

a. Provide a datasetId for your new dataset. This guide uses myFactsQaDataset.

b. Select Flow dataset type.

c. Leave the validation target field empty and click Save

Your new dataset page appears, showing an empty dataset. Add examples to it by following these steps:

a. Click the Add example button to open the example editor panel.

b. Only the Input field is required. Enter "Who is man's best friend?" in the Input field, and click Save to add the example has to your dataset.

If you have configured the EvaluatorRegex metric and would like to try it out, you need to specify a Reference string that contains the pattern to match the output against. For the preceding input, set the Reference output text to "(?i)dog", which is a case-insensitive regular- expression pattern to match the word "dog" in the flow output.

c. Repeat steps (a) and (b) a couple of more times to add more examples. This guide adds the following example inputs to the dataset:


"Can I give milk to my cats?"
"From which animals did dogs evolve?"
If you are using the regular-expression evaluator, use the corresponding reference strings:


"(?i)don't know"
"(?i)wolf|wolves"
Note that this is a contrived example and the regular-expression evaluator may not be the right choice to evaluate the responses from qaFlow. However, this guide can be applied to any Genkit Go evaluator of your choice.

By the end of this step, your dataset should have 3 examples in it, with the values mentioned above.

Run evaluation and view results
To start evaluating the flow, click the Run new evaluation button on your dataset page. You can also start a new evaluation from the Evaluations tab.

Select the Flow radio button to evaluate a flow.

Select qaFlow as the target flow to evaluate.

Select myFactsQaDataset as the target dataset to use for evaluation.

If you have installed an evaluator metric using Genkit plugins, you can see these metrics in this page. Select the metrics that you want to use with this evaluation run. This is entirely optional: Omitting this step will still return the results in the evaluation run, but without any associated metrics.

If you have not provided any reference values and are using the EvaluatorRegex metric, your evaluation will fail since this metric needs reference to be set.

Click Run evaluation to start evaluation. Depending on the flow you're testing, this may take a while. Once the evaluation is complete, a success message appears with a link to view the results. Click the link to go to the Evaluation details page.

You can see the details of your evaluation on this page, including original input, extracted context and metrics (if any).

Core concepts
Terminology
Knowing the following terms can help ensure that you correctly understand the information provided on this page:

Evaluation: An evaluation is a process that assesses system performance. In Genkit, such a system is usually a Genkit primitive, such as a flow or a model. An evaluation can be automated or manual (human evaluation).

Bulk inference Inference is the act of running an input on a flow or model to get the corresponding output. Bulk inference involves performing inference on multiple inputs simultaneously.

Metric An evaluation metric is a criterion on which an inference is scored. Examples include accuracy, faithfulness, maliciousness, whether the output is in English, etc.

Dataset A dataset is a collection of examples to use for inference-based evaluation. A dataset typically consists of Input and optional Reference fields. The Reference field does not affect the inference step of evaluation but it is passed verbatim to any evaluation metrics. In Genkit, you can create a dataset through the Dev UI. There are two types of datasets in Genkit: Flow datasets and Model datasets.

Supported evaluators
Genkit supports several evaluators, some built-in, and others provided externally.

Genkit evaluators
Genkit includes a small number of built-in evaluators, ported from the JS evaluators plugin, to help you get started:

EvaluatorDeepEqual -- Checks if the generated output is deep-equal to the reference output provided.
EvaluatorRegex -- Checks if the generated output matches the regular expression provided in the reference field.
EvaluatorJsonata -- Checks if the generated output matches the JSONATA expression provided in the reference field.
Advanced use
Along with its basic functionality, Genkit also provides advanced support for certain evaluation use cases.

Evaluation using the CLI
Genkit CLI provides a rich API for performing evaluation. This is especially useful in environments where the Dev UI is not available (e.g. in a CI/CD workflow).

Genkit CLI provides 3 main evaluation commands: eval:flow, eval:extractData, and eval:run.

Evaluation eval:flow command
The eval:flow command runs inference-based evaluation on an input dataset. This dataset may be provided either as a JSON file or by referencing an existing dataset in your Genkit runtime.


# Referencing an existing dataset
genkit eval:flow qaFlow --input myFactsQaDataset
# or, using a dataset from a file
genkit eval:flow qaFlow --input testInputs.json
Note: Make sure that you start your genkit app before running these CLI commands.


genkit start -- go run main.go
Here, testInputs.json should be an array of objects containing an input field and an optional reference field, like below:


[
  {
    "input": "What is the French word for Cheese?"
  },
  {
    "input": "What green vegetable looks like cauliflower?",
    "reference": "Broccoli"
  }
]
If your flow requires auth, you may specify it using the --context argument:


genkit eval:flow qaFlow --input testInputs.json --context '{"auth": {"email_verified": true}}'
By default, the eval:flow and eval:run commands use all available metrics for evaluation. To run on a subset of the configured evaluators, use the --evaluators flag and provide a comma-separated list of evaluators by name:


genkit eval:flow qaFlow --input testInputs.json --evaluators=genkitEval/regex,genkitEval/jsonata
You can view the results of your evaluation run in the Dev UI at localhost:4000/evaluate.

eval:extractData and eval:run commands
To support raw evaluation, Genkit provides tools to extract data from traces and run evaluation metrics on extracted data. This is useful, for example, if you are using a different framework for evaluation or if you are collecting inferences from a different environment to test locally for output quality.

You can batch run your Genkit flow and extract an evaluation dataset from the resultant traces. A raw evaluation dataset is a collection of inputs for evaluation metrics, without running any prior inference.

Run your flow over your test inputs:


genkit flow:batchRun qaFlow testInputs.json
Extract the evaluation data:


genkit eval:extractData qaFlow --maxRows 2 --output factsEvalDataset.json
The exported data has a format different from the dataset format presented earlier. This is because this data is intended to be used with evaluation metrics directly, without any inference step. Here is the syntax of the extracted data.


Array<{
  "testCaseId": string,
  "input": any,
  "output": any,
  "context": any[],
  "traceIds": string[],
}>;
The data extractor automatically locates retrievers and adds the produced docs to the context array. You can run evaluation metrics on this extracted dataset using the eval:run command.


genkit eval:run factsEvalDataset.json
By default, eval:run runs against all configured evaluators, and as with eval:flow, results for eval:run appear in the evaluation page of Developer UI, located at localhost:4000/evaluate.

---

Monitoring 

bookmark_border
Genkit provides two complementary monitoring features: OpenTelemetry export and trace inspection using the developer UI.

OpenTelemetry export
Genkit is fully instrumented with OpenTelemetry and provides hooks to export telemetry data.

The Google Cloud plugin exports telemetry to Cloud's operations suite.

Trace store
The trace store feature is complementary to the telemetry instrumentation. It lets you inspect your traces for your flow runs in the Genkit Developer UI.

This feature is enabled whenever you run a Genkit flow in a dev environment (such as when using genkit start or genkit flow:run).

---

Writing Genkit plugins 

bookmark_border
Genkit's capabilities are designed to be extended by plugins. Genkit plugins are configurable modules that can provide models, retrievers, indexers, trace stores, and more. You've already seen plugins in action just by using Genkit:


import (
    "github.com/firebase/genkit/go/ai"
    "github.com/firebase/genkit/go/genkit"
    "github.com/firebase/genkit/go/plugins/googlegenai"
    "github.com/firebase/genkit/go/plugins/server"
)

g, err := genkit.Init(ctx,
    ai.WithPlugins(
        &googlegenai.GoogleAI{APIKey: ...},
        &googlegenai.VertexAI{ProjectID: "my-project", Location: "us-central1"},
    ),
)
The Vertex AI plugin takes configuration (such as the user's Google Cloud project ID) and registers a variety of new models, embedders, and more with the Genkit registry. The registry serves as a lookup service for named actions at runtime, and powers Genkit's local UI for running and inspecting models, prompts, and more.

Creating a plugin
In Go, a Genkit plugin is a package that adheres to a small set of conventions. A single module can contain several plugins.

Provider ID
Every plugin must have a unique identifier string that distinguishes it from other plugins. Genkit uses this identifier as a namespace for every resource your plugin defines, to prevent naming conflicts with other plugins.

For example, if your plugin has an ID yourplugin and provides a model called text-generator, the full model identifier will be yourplugin/text-generator.

This provider ID needs to be exported and you should define it once for your plugin and use it consistently when required by a Genkit function.


const providerID = "yourplugin"
Standard exports
Every plugin should define and export the following symbols to conform to the genkit.Plugin interface:

A struct type that encapsulates all of the configuration options accepted by the plugin.

For any plugin options that are secret values, such as API keys, you should offer both a config option and a default environment variable to configure it. This lets your plugin take advantage of the secret-management features offered by many hosting providers (such as Cloud Secret Manager, which you can use with Cloud Run). For example:


type MyPlugin struct {
    APIKey string
    // Other options you may allow to configure...
}
A Name() method on the struct that returns the provider ID.

An Init() method on the struct with a declaration like the following:


func (m *MyPlugin) Init(ctx context.Context, g *genkit.Genkit) error
In this function, perform any setup steps required by your plugin. For example:

Confirm that any required configuration values are specified and assign default values to any unspecified optional settings.
Verify that the given configuration options are valid together.
Create any shared resources required by the rest of your plugin. For example, create clients for any services your plugin accesses.
To the extent possible, the resources provided by your plugin shouldn't assume that any other plugins have been installed before this one.

This method will be called automatically during genkit.Init() when the user passes the plugin into the WithPlugins() option.

Building plugin features
A single plugin can activate many new things within Genkit. For example, the Vertex AI plugin activates several new models as well as an embedder.

Model plugins
Genkit model plugins add one or more generative AI models to the Genkit registry. A model represents any generative model that is capable of receiving a prompt as input and generating text, media, or data as output.

See Writing a Genkit model plugin.

Telemetry plugins
Genkit telemetry plugins configure Genkit's OpenTelemetry instrumentation to export traces, metrics, and logs to a particular monitoring or visualization tool.

See Writing a Genkit telemetry plugin.

Publishing a plugin
Genkit plugins can be published as normal Go packages. To increase discoverability, your package should have genkit somewhere in its name so it can be found with a simple search on pkg.go.dev. Any of the following are good choices:

github.com/yourorg/genkit-plugins/servicename
github.com/yourorg/your-repo/genkit/servicename

---

Writing a Genkit model plugin 

bookmark_border
Genkit model plugins add one or more generative AI models to the Genkit registry. A model represents any generative model that is capable of receiving a prompt as input and generating text, media, or data as output.

Before you begin
Read Writing Genkit plugins for information about writing any kind of Genkit plug-in, including model plugins. In particular, note that every plugin must export a type that conforms to the genkit.Plugin interface, which includes a Name() and a Init() function.

Model definitions
Generally, a model plugin will make one or more genkit.DefineModel() calls in its Init function—once for each model the plugin is providing an interface to.

A model definition consists of three components:

Metadata declaring the model's capabilities.
A configuration type with any specific parameters supported by the model.
A generation function that accepts an ai.ModelRequest and returns an ai.ModelResponse, presumably using an AI model to generate the latter.
At a high level, here's what it looks like in code:


type MyModelConfig struct {
    ai.GenerationCommonConfig
    AnotherCustomOption string
    CustomOption        int
}

DefineModel(g, providerID, "my-model",
    &ai.ModelInfo{
        Label: "My Model",
        Supports: &ai.ModelSupports{
            Multiturn:  true,  // Does the model support multi-turn chats?
            SystemRole: true,  // Does the model support syatem messages?
            Media:      false, // Can the model accept media input?
            Tools:      false, // Does the model support function calling (tools)?
        },
        Versions: []string{"my-model-001", "..."},
    },
    func(ctx context.Context, mr *ai.ModelRequest, _ ai.ModelStreamCallback) (*ai.ModelResponse, error) {
        // Verify that the request includes a configuration that conforms to your schema.
        if _, ok := mr.Config.(MyModelConfig); !ok {
            return nil, fmt.Errorf("request config must be type MyModelConfig")
        }

        // Use your custom logic to convert Genkit's ai.ModelRequest into a form
        // usable by the model's native API.
        apiRequest, err := apiRequestFromGenkitRequest(genRequest)
        if err != nil {
            return nil, err
        }

        // Send the request to the model API, using your own code or the model
        // API's client library.
        apiResponse, err := callModelAPI(apiRequest)
        if err != nil {
            return nil, err
        }

        // Use your custom logic to convert the model's response to Genkin's ai.ModelResponse.
        response, err := genResponseFromAPIResponse(apiResponse)
        if err != nil {
            return nil, err
        }

        return response, nil
    },
)
Declaring model capabilities
Every model definition must contain, as part of its metadata, an ai.ModelInfo value that declares which features the model supports. Genkit uses this information to determine certain behaviors, such as verifying whether certain inputs are valid for the model. For example, if the model doesn't support multi-turn interactions, then it's an error to pass it a message history.

Note that these declarations refer to the capabilities of the model as provided by your plugin, and do not necessarily map one-to-one to the capabilities of the underlying model and model API. For example, even if the model API doesn't provide a specific way to define system messages, your plugin might still declare support for the system role, and implement it as special logic that inserts system messages into the user prompt.

Defining your model's config schema
To specify the generation options a model supports, define and export a configuration type. Genkit has an ai.GenerationCommonConfig type that contains options frequently supported by generative AI model services, which you can embed or use outright.

Your generation function should verify that the request contains the correct options type.

Transforming requests and responses
The generation function carries out the primary work of a Genkit model plugin: transforming the ai.ModelRequest from Genkit's common format into a format that is supported by your model's API, and then transforming the response from your model into the ai.ModelResponse format used by Genkit.

Sometimes, this may require massaging or manipulating data to work around model limitations. For example, if your model does not natively support a system message, you may need to transform a prompt's system message into a user-model message pair.

Exports
In addition to the resources that all plugins must export, a model plugin should also export the following:

A generation config type, as discussed earlier.

A Model() function, which returns references to your plugin's defined models. Often, this can be:


func Model(g *genkit.Genkit, name string) *ai.Model {
    return genkit.LookupModel(g, providerID, name)
}
A ModelRef function, which creates a model reference paired with its config that can validate the type and be passed around together:


func ModelRef(name string, config *MyModelConfig) *ai.ModelRef {
    return ai.NewModelRef(name, config)
}
Optional: A DefineModel() function, which lets users define models that your plugin can provide, but that you do not automatically define. There are two main reasons why you might want to provide such a function:

Your plugin provides access to too many models to practically register each one. For example, the Ollama plugin can provide access to dozens of different models, with more added frequently. For this reason, it doesn't automatically define any models, and instead requires the user to call DefineModel() for each model they want to use.

To give your users the ability to use newly-released models that you have not yet added to your plugin.

A plugin's DefineModel() function is typically a frontend to genkit.DefineModel() that defines a generation function, but lets the user specify the model name and model capabilities.

---

Writing a Genkit telemetry plugin 

bookmark_border


The Genkit libraries are instrumented with OpenTelemetry to support collecting traces, metrics, and logs. Genkit users can export this telemetry data to monitoring and visualization tools by installing a plugin that configures the OpenTelemetry Go SDK to export to a particular OpenTelemetry-capable system.

Genkit includes a plugin that configures OpenTelemetry to export data to Google Cloud Monitoring and Cloud Logging. To support other monitoring systems, you can extend Genkit by writing a telemetry plugin, as described on this page.

Before you begin
Read Writing Genkit plugins for information about writing any kind of Genkit plugin, including telemetry plugins. In particular, note that every plugin must export an Init function, which users are expected to call before using the plugin.

Exporters and Loggers
As stated earlier, the primary job of a telemetry plugin is to configure OpenTelemetry (which Genkit has already been instrumented with) to export data to a particular service. To do so, you need the following:

An implementation of OpenTelemetry's SpanExporter interface that exports data to the service of your choice.
An implementation of OpenTelemetry's metric.Exporter interface that exports data to the service of your choice.
Either a slog.Logger or an implementation of the slog.Handler interface, that exports logs to the service of your choice.
Depending on the service you're interested in exporting to, this might be a relatively minor effort or a large one.

Because OpenTelemetry is an industry standard, many monitoring services already have libraries that implement these interfaces. For example, the googlecloud plugin for Genkit makes use of the opentelemetry-operations-go library, maintained by the Google Cloud team. Similarly, many monitoring services provide libraries that implement the standard slog interfaces.

On the other hand, if no such libraries are available for your service, implementing the necessary interfaces can be a substantial project.

Check the OpenTelemetry registry or the monitoring service's docs to see if integrations are already available.

If you need to build these integrations yourself, take a look at the source of the official OpenTelemetry exporters and the page A Guide to Writing slog Handlers.

Building the plugin
Dependencies
Every telemetry plugin needs to import the Genkit core library and several OpenTelemetry libraries:


import {
	// Import the Genkit core library.

	"github.com/firebase/genkit/go/genkit"

	// Import the OpenTelemetry libraries.
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/trace"
}
If you are building a plugin around an existing OpenTelemetry or slog integration, you will also need to import them.

Config
A telemetry plugin should, at a minimum, support the following configuration options:


type Config struct {
	// Export even in the dev environment.
	ForceExport bool

	// The interval for exporting metric data.
	// The default is 60 seconds.
	MetricInterval time.Duration

	// The minimum level at which logs will be written.
	// Defaults to [slog.LevelInfo].
	LogLevel slog.Leveler
}

The examples that follow assume you are making these options available and will provide some guidance on how to handle them.

Most plugins will also include configuration settings for the service it's exporting to (API key, project name, and so on).

Init()
The Init() function of a telemetry plugin should do all of the following:

Return early if Genkit is running in a development environment (such as when running with with genkit start) and the Config.ForceExport option isn't set:


shouldExport := cfg.ForceExport || os.Getenv("GENKIT_ENV") != "dev"
if !shouldExport {
	return nil
}
Initialize your trace span exporter and register it with Genkit:


spanProcessor := trace.NewBatchSpanProcessor(YourCustomSpanExporter{})
genkit.RegisterSpanProcessor(g, spanProcessor)
Initialize your metric exporter and register it with the OpenTelemetry library:


r := metric.NewPeriodicReader(
	YourCustomMetricExporter{},
	metric.WithInterval(cfg.MetricInterval),
)
mp := metric.NewMeterProvider(metric.WithReader(r))
otel.SetMeterProvider(mp)
Use the user-configured collection interval (Config.MetricInterval) when initializing the PeriodicReader.

Register your slog handler as the default logger:


logger := slog.New(YourCustomHandler{
	Options: &slog.HandlerOptions{Level: cfg.LogLevel},
})
slog.SetDefault(logger)
You should configure your handler to honor the user-specified minimum log level (Config.LogLevel).

PII redaction
Because most generative AI flows begin with user input of some kind, it's a likely possibility that some flow traces contain personally-identifiable information (PII). To protect your users' information, you should redact PII from traces before you export them.

If you are building your own span exporter, you can build this functionality into it.

If you're building your plugin around an existing OpenTelemetry integration, you can wrap the provided span exporter with a custom exporter that carries out this task. For example, the googlecloud plugin removes the genkit:input and genkit:output attributes from every span before exporting them using a wrapper similar to the following:


type redactingSpanExporter struct {
	trace.SpanExporter
}

func (e *redactingSpanExporter) ExportSpans(ctx context.Context, spanData []trace.ReadOnlySpan) error {
	var redacted []trace.ReadOnlySpan
	for _, s := range spanData {
		redacted = append(redacted, redactedSpan{s})
	}
	return e.SpanExporter.ExportSpans(ctx, redacted)
}

func (e *redactingSpanExporter) Shutdown(ctx context.Context) error {
	return e.SpanExporter.Shutdown(ctx)
}

type redactedSpan struct {
	trace.ReadOnlySpan
}

func (s redactedSpan) Attributes() []attribute.KeyValue {
	// Omit input and output, which may contain PII.
	var ts []attribute.KeyValue
	for _, a := range s.ReadOnlySpan.Attributes() {
		if a.Key == "genkit:input" || a.Key == "genkit:output" {
			continue
		}
		ts = append(ts, a)
	}
	return ts
}

Troubleshooting
If you're having trouble getting data to show up where you expect, OpenTelemetry provides a useful diagnostic tool that helps locate the source of the problem.

---


# morphs

Morphs are WebAssembly Components that are deployable on the Hayride platform. 

This repository contains a number of community driven morphs and examples of how to build them.

## Getting Started  

The majority of the morphs are written in TinyGo/Go. 

To get started, you will need to install various WebAssembly tools. You can learn more about the various ecosystem tools in the [Hayride documentation](https://hayride.dev/docs/).

## Components 

The majority of the components in this repository are exported implementations of the Hayride WebAssembly Interfaces (WIT) defined in the [hayride-dev/coven](https://github.com/hayride-dev/coven).

### Example Component

AI Agents is an example of a reactor component that implements the `hayride:ai/agents` interface.

```
package hayride:agents@0.0.1;

world default {
    include hayride:wasip2/imports@0.0.63;
    export hayride:ai/agents@0.0.63;
}
```

Once the component is built, it can be used in a composition to create a more complex functionality.

## Compositions 

Morphs can be composed together to create more complex functionality. This allows developers to build upon existing morphs and create new ones by combining their capabilities. Compositions can be defined using [wac](https://github.com/bytecodealliance/wac), a tool for composing WebAssembly components.


### Example Composition 

Hayride defines WebAssembly Interfaces in [coven](https://github.com/hayride-dev/coven).

Taking a look at the WIT definitions for the cli AI `runner`: 
```
world cli {
    include hayride:wasip2/imports@0.0.63;
    include hayride:wasip2/exports@0.0.63;
    
    import hayride:ai/agents@0.0.63;
    import hayride:ai/model-repository@0.0.63;
}
```

We can see that we are importing the `hayride:ai/agents` and `hayride:ai/model-repository` interfaces.

`hayride:ai/agents` itself, imports `hayride:ai/tools`, `hayride:ai/models`, and `hayride:ai/context` interfaces.

This means that the `cli` composition can use the `agents`, `tools`, and `models` interfaces to create a command line interface for interacting with AI agents.

You can find exported implementations of these interfaces in the components/ai directory of this repository.

Using these components, we can satisfy the `cli` imports by creating a composition that includes the necessary components:

```
package hayride:agent;

let context = new hayride:inmemory@0.0.1 {...}; 

let llama = new hayride:llama31@0.0.1 {...};

let tools = new hayride:default-tools@0.0.1 {...};

let agent = new hayride:default-agent@0.0.1 {
  context: context.context,
  model: llama.model,
  tools: tools.tools,
  ...
};

let cli = new hayride:cli@0.0.1 {
  context: context.context,
  model: llama.model,
  tools: tools.tools,
  agents: agent.agents,
  ...
};

// Export the cli
export cli...;
```

This composition creates a command line interface that can interact with AI agents using the `hayride:ai/agents` and host implementation of the `hayride:ai/model-repository` interfaces.

## Community Morphs

We welcome contributions from the community! If you have a morph that you'd like to share, please submit a pull request.

Currently, the easiest morph to develop is the `hayride:ai/model` morph, which provides LLM specific token formatting and encoding. Adding model support is a great way to get started with morph development.

## Contributing
Contributions are welcome! If you'd like to contribute, please follow these steps:

- Fork the repository.
- Create a new branch for your feature or bug fix.
- Submit a pull request with a detailed description of your changes.

## License
This project is licensed under the MIT License. See the LICENSE file for details
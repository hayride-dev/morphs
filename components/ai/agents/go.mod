module github.com/hayride-dev/morphs/components/ai/agents

go 1.23.6

replace github.com/hayride-dev/morphs/components/ai/tools/datetime => ../tools/datetime

replace github.com/hayride-dev/bindings => ../../../../bindings

require (
	github.com/hayride-dev/bindings v0.0.21
	github.com/hayride-dev/morphs/components/ai/tools/datetime v0.0.0
)

require go.bytecodealliance.org/cm v0.2.2 // indirect

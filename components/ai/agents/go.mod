module github.com/hayride-dev/morphs/components/ai/agents

go 1.23.6

replace github.com/hayride-dev/morphs/components/ai/tools/datetime => ../tools/datetime

require (
	github.com/hayride-dev/bindings v0.0.33
	github.com/hayride-dev/morphs/components/ai/tools/datetime v0.0.0-00010101000000-000000000000
	go.bytecodealliance.org/cm v0.2.2
)

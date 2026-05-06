# jolt-go

Go binding for [JoltPhysics](https://github.com/jrouwe/JoltPhysics) via the [`amerkoleci/joltc`](https://github.com/amerkoleci/joltc) C binding.

**Status:** MVP under construction.

Module: `github.com/gataengine/jolt-go`

## License

MIT. See [LICENSE](./LICENSE).

## Pinned upstream

- `amerkoleci/joltc` @ `52d8c98df523f449eb3e01b1060a0fde052970d1` (tracks Jolt `v5.5.0`)
- `jrouwe/JoltPhysics` @ `23dadd0e603f1b321142d4c74df07fce85064989` (tag `v5.5.0`)

To regenerate bindings after bumping submodules: `go generate ./jolt`.

# Yōk
Yōk is a scripting language designed to replace bash scripts for simple use cases.
Yōk scripts can be transpiled directly into bash preventing the need to install an interpreter on your target system. The language also suports modern lanagues features including:
- Modernized syntax
- An auto format command
- Increase focus on readability over tersness
- A basic type system
- Improved error handling
- Built-in test support

Yōk is currently in active development so not all its features are fully implemented.
Additionally, many of the existing features are subject to change.

## TODO
The following tasks should be completed in the near future:
- [ ] The Lex phase should produce a stream of parse.Tokens rather than Nodes.
- [x] Fix the double-wrapping behavior occuring when building the AST.
- [ ] Seperate Expr and Stmt ast.builders to make building if, call, and other nodes more robust.

The following tasks are still in progress:
- [ ] Add a yok run command to complie and then run yok scripts in the same step.
- [ ] Add a validation phase to the AST.
- [ ] Design a sutable error package.
    - [ ] How should internal complier errors be handled.
    - [ ] How should user facing errors be displayed (hoping to borrow from rust for this).
- [ ] Add in the basics of the type system.
- [ ] Add support arrays.
- [ ] Add support tables (associative arrasy in bash).
- [ ] Add support for looping constructs.
- [ ] Add support for an error type (look at TS and Rust types).
    - [ ] this should include designing good methods of handling (unwraping?) errors.
- [ ] Add support for functions.
- [ ] Build a language reference.
- [ ] Add the 'source' keyword for brining in external files.
- [ ] Design a testing system so that itteration on scripts can move faster.
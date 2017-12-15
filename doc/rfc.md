# Draft Machines Specification RFC

## Overview

A machine consists of 

1. A current set of bindings,
1. The name of current node, and
1. A (pointer to a) machine specification.

The top-level function (`run`) takes a machine and a message as inputs
and outputs

1. A new set of bindings,
2. A new node name, and
3. Zero or more messages as output.


## Notation

1.  A _type name_ is a symbol representing a type.  In this document a
    type name is a capitalized symbol formatted like `This`.

1.  An all-uppercase symbol represents a type variable.  For example,
    `TYPE` represents any given type.

1.  The type `Nothing` is hard to define but it somehow means what it
    says.

1.  A type variable of the form `TYPE?` represents the type that's
    either `TYPE` or `Nothing`.

1.  A `List<TYPE>` is an ordered finite set of elements of type `TYPE`.

1. `Set<TYPE>` is the type of a set that contains elements of type
   `TYPE`.

1. `Cross<TYPE1,...,TYPEN>` is the cross product of those types.

1.  A `Pair<LEFT,RIGHT>` is a `CROSS<LEFT,RIGHT>`.  The left side of a
    pair is called a _key_, and the right side is called a _value_.
    Wild, I know.
   
1.  A `Map<KEYTYPE,VALTYPE>` is a `Set<Pair<KEYTYPE,VALTYPE>>` such
    that each key is unique.

1.  To specify a structure, we use the following representation:

    ```
    {key1: TYPE1, ..., keyn: TYPEN}
    ```

    with the obvious interpretation.  A `key:TYPE` in a structure is
    called a _component_.  A key is represented as a lower-case symbol
    like `this`.


## Messages and patterns

1. `Number` represents a number.

1. `String` represents a string.

1.  A `Expression` is either a `Number`, `String`,
    `Map<String,Expression>`, or a `Set` of one of those types.
	
1.  A `Variable` is a `String`. (In practice, by convention a
    `Variable` is a string that starts with the character `?`.)
	
1.  A `Message` is either a `Number`, `String`, `Map<String,Message>`,
    or a `Set` of one of those types.
	
1.  A `Pattern` is either a `Variable`, `Number`, `String`,
    `Map<Variable,Pattern>`, or a `Set` of one of those types.
	
1.  A `Bindings` is a `Map<String,Message>`.


## Pattern matching

`match` is a function from `Cross<Pattern,Message,Bindings>` to
`Set<Bindings>`.

Constraints ToDo.


## Machine specifications

1.  A `Specification` is a structure with the following components:

    ```
	name: String
	nodes: Map<NodeName,Node>
	```

1.  A `NodeName` is a string.

1.  A `Node` is a structure will the following components:

    ```
	type: Nodetype?
	action: Action?
	branching: Branching?
	```
	
	`type` defaults to `"bindings"`

1. `Nodetype` is either the literal `"message"` or `"bindings"`.

1.  An `Action` is a structure with the following components:

    ```
	interpreter: Interpreter
	source: Source
	```

    Technically the type `Source` depends on the type `Interpreter`,
    but we'll not head that way (for now).
	
	Practically speaking, an `Interpreter` is often just a string
    (perhaps a URL) that can be resolved to something that knows how
    to compile (optionally) and execute some `Source` that's just a
    `String`.
	
	Also in practice, an `Action` might have other components that the
    `Interpreter` knows how to use.
	
1.  An `Interpreter` is a function from `Source` to `Execution`.	

1.  An `Execution` is a structure that contains at least the following
    components:
	
	```
	bindings: Bindings
	emitted: Set<Message>
	```

1.  A `Branching` is a `List<Branch>`.

1.  A `Branch` is a structure with the following components:

    ```
	pattern: Pattern?
	guard: Guard?
	target: NodeName
	```

1.  A `Guard` is (for now) an `Action` that does not emit any messages
    (`Execution.emitted`) when executed.

## Machines

1.  A `Machine` is a structure with the following components:

    ```
	nodeName: NodeName
	bindings: Bindings
	spec: Spec
	```
	
	In practice, the `spec` is a string that can be resolved to a
    `Spec` or a pointer to a `Spec`.

## Message processing

`run` is a function from `Cross<Machine,Message>` to `Ran`.

1.  `Ran` is a structure with (at least) the following
    components:
   
    ```
	nodeName: NodeName
	bindings: Bindings
	emitted: Set<Message>
	```
	
	In practice, this structure can include other data such as trace
    records that include intermediate states.

Constraints ToDo!


## Refinements and extensions

1.  `run` can also include a `Control` argument.

2.  `Control` is a structure with the following components:

    ```
	stepLimit: NaturalNumber
	```
	
	More ToDo.

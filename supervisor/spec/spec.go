package spec

type Ref struct {}

type Spec interface {
	ChildSpec() Spec
}

type SpecsMap map[string]Spec


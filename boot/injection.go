package boot

type Injector interface {
	BuildDependencies() (singletons []interface{}, err error)
}

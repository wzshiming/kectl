package handle

type ObjectRef struct {
	Name      string
	Namespace string
}

type KMetadata interface {
	GetName() string
	GetNamespace() string
}

// KObj returns ObjectRef from ObjectMeta
func KObj[T KMetadata](obj T) ObjectRef {
	return ObjectRef{
		Name:      obj.GetName(),
		Namespace: obj.GetNamespace(),
	}
}

// KRef returns ObjectRef from name and namespace
func KRef(namespace, name string) ObjectRef {
	return ObjectRef{
		Name:      name,
		Namespace: namespace,
	}
}

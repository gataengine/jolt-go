package generator

// Strategy selects how a settings struct is emitted.
type Strategy int

const (
	StrategyDirect    Strategy = iota // Approach 1 — layout-match
	StrategyTranslate                 // Approach 2 — field-by-field
)

// ListenerSpec describes a JPH_*Listener vtable that the generator emits as
// a struct-of-funcs Go API.
type ListenerSpec struct {
	CType     string // e.g., "JPH_ContactListener"
	ProcsType string // e.g., "JPH_ContactListener_Procs"
	SetProcs  string // e.g., "JPH_ContactListener_SetProcs"
	Methods   []ListenerMethod
}

// ListenerMethod is a single callback in a listener vtable.
type ListenerMethod struct {
	Name    string
	Returns string // empty for void
}

// Sidecar covers the things joltc.h alone can't express.
type Sidecar struct {
	Inheritance      map[string]string   // child Go type -> direct parent Go type
	RefCounted       []string            // Go types using AddRef/Release
	Listeners        []ListenerSpec
	SettingsInit     map[string]string   // settings type -> JPH_*_Init function
	SettingsStrategy map[string]Strategy // override per type
}

// Annotations is the MVP sidecar populated for Phase 8 codegen runs.
// Add entries as MVP scope expands.
var Annotations = Sidecar{
	Inheritance: map[string]string{
		"BoxShape":                "ConvexShape",
		"SphereShape":             "ConvexShape",
		"CapsuleShape":            "ConvexShape",
		"TaperedCapsuleShape":     "ConvexShape",
		"CylinderShape":           "ConvexShape",
		"TaperedCylinderShape":    "ConvexShape",
		"ConvexHullShape":         "ConvexShape",
		"TriangleShape":           "ConvexShape",
		"PlaneShape":              "ConvexShape",
		"EmptyShape":              "Shape",
		"MeshShape":               "Shape",
		"HeightFieldShape":        "Shape",
		"StaticCompoundShape":     "CompoundShape",
		"MutableCompoundShape":    "CompoundShape",
		"CompoundShape":           "Shape",
		"RotatedTranslatedShape":  "DecoratedShape",
		"ScaledShape":             "DecoratedShape",
		"OffsetCenterOfMassShape": "DecoratedShape",
		"DecoratedShape":          "Shape",
		"ConvexShape":             "Shape",
	},
	RefCounted: []string{
		"Shape", "ConvexShape", "CompoundShape", "DecoratedShape",
		"BoxShape", "SphereShape", "CapsuleShape", "TaperedCapsuleShape",
		"CylinderShape", "TaperedCylinderShape", "ConvexHullShape",
		"TriangleShape", "PlaneShape", "EmptyShape", "MeshShape",
		"HeightFieldShape", "StaticCompoundShape", "MutableCompoundShape",
		"RotatedTranslatedShape", "ScaledShape", "OffsetCenterOfMassShape",
		"PhysicsMaterial",
	},
	Listeners: []ListenerSpec{{
		CType:     "JPH_ContactListener",
		ProcsType: "JPH_ContactListener_Procs",
		SetProcs:  "JPH_ContactListener_SetProcs",
		Methods: []ListenerMethod{
			{Name: "OnContactValidate", Returns: "JPH_ValidateResult"},
			{Name: "OnContactAdded"},
			{Name: "OnContactPersisted"},
			{Name: "OnContactRemoved"},
		},
	}},
	SettingsInit: map[string]string{
		"BodyCreationSettings": "JPH_BodyCreationSettings_Init",
	},
	SettingsStrategy: map[string]Strategy{},
}

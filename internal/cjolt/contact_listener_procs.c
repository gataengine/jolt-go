// contact_listener_procs.c — registers the Go //export trampolines with
// JPH_ContactListener_SetProcs. This is a plain C file (not a cgo preamble)
// so cgo automatically places _cgo_export.h (which declares the Go-exported
// functions as extern C prototypes) on the include path when the package is
// built. The Go init() in cjolt_generated.go calls register_contact_listener_procs
// via cgo once at program startup.
//
// The Go trampolines do not declare their struct-pointer params as const (cgo
// generates non-const prototypes for //export functions). The struct fields in
// JPH_ContactListener_Procs expect const pointers, so we cast explicitly.

#include "joltc.h"
#include "_cgo_export.h"

/* s_contact_listener_procs must have static storage duration: JPH_ContactListener_SetProcs
   stores a pointer to it, not a copy, so the struct must outlive the call. */
static JPH_ContactListener_Procs s_contact_listener_procs;

void register_contact_listener_procs(void) {
	s_contact_listener_procs.OnContactValidate  = (JPH_ValidateResult(*)(void*, const JPH_Body*, const JPH_Body*, const JPH_RVec3*, const JPH_CollideShapeResult*))jolt_go_contactListener_OnContactValidate;
	s_contact_listener_procs.OnContactAdded     = (void(*)(void*, const JPH_Body*, const JPH_Body*, const JPH_ContactManifold*, JPH_ContactSettings*))jolt_go_contactListener_OnContactAdded;
	s_contact_listener_procs.OnContactPersisted = (void(*)(void*, const JPH_Body*, const JPH_Body*, const JPH_ContactManifold*, JPH_ContactSettings*))jolt_go_contactListener_OnContactPersisted;
	s_contact_listener_procs.OnContactRemoved   = (void(*)(void*, const JPH_SubShapeIDPair*))jolt_go_contactListener_OnContactRemoved;
	JPH_ContactListener_SetProcs(&s_contact_listener_procs);
}

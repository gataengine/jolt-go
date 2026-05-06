#ifndef FUNCS_H
#define FUNCS_H

#include <stdbool.h>
#include <stdint.h>

typedef struct JPH_Body JPH_Body;
typedef struct JPH_Vec3 { float x, y, z; } JPH_Vec3;

void JPH_RegisterTypes(void);
void JPH_UnregisterTypes(void);
JPH_Vec3 JPH_Body_GetPosition(const JPH_Body* self);
void JPH_Body_SetLinearVelocity(JPH_Body* self, const JPH_Vec3* v);

#endif

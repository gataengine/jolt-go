#ifndef SETTINGS_H
#define SETTINGS_H
#include <stdint.h>
#include <stdbool.h>

typedef struct JPH_Vec3 { float x, y, z; } JPH_Vec3;
typedef struct JPH_Quat { float x, y, z, w; } JPH_Quat;
typedef struct JPH_RVec3 { float x, y, z; } JPH_RVec3;
typedef struct JPH_Shape JPH_Shape;
typedef uint16_t JPH_ObjectLayer;
typedef uint8_t JPH_MotionType;

/* JPH_BodyCreationSettings is opaque in joltc — exposed only via
   Create/Destroy and accessor methods (no public field layout). */
typedef struct JPH_BodyCreationSettings JPH_BodyCreationSettings;

JPH_BodyCreationSettings* JPH_BodyCreationSettings_Create(void);
void JPH_BodyCreationSettings_Destroy(JPH_BodyCreationSettings*);

float JPH_BodyCreationSettings_GetFriction(const JPH_BodyCreationSettings*);
void  JPH_BodyCreationSettings_SetFriction(JPH_BodyCreationSettings*, float);
float JPH_BodyCreationSettings_GetLinearDamping(const JPH_BodyCreationSettings*);
void  JPH_BodyCreationSettings_SetLinearDamping(JPH_BodyCreationSettings*, float);

#endif

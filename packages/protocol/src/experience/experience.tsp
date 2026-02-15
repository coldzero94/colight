import "@typespec/http";
import "../common/errors.tsp";

using TypeSpec.Http;

namespace Colight.Experience;

// ─── Models ───

model ExperienceWeapon {
  id: string;

  @encodedName("application/json", "weapon_code")
  weaponCode: string;

  confidence: float32;

  @encodedName("application/json", "is_primary")
  isPrimary: boolean;

  reasoning: string;

  @encodedName("application/json", "user_confirmed")
  userConfirmed: boolean;

  @encodedName("application/json", "user_modified")
  userModified: boolean;
}

model Experience {
  id: string;

  @encodedName("application/json", "user_id")
  userId: string;

  title: string;
  category: string;

  @encodedName("application/json", "period_start")
  periodStart?: string;

  @encodedName("application/json", "period_end")
  periodEnd?: string;

  role: string;
  content: string;
  result: string;

  @encodedName("application/json", "star_situation")
  starSituation: string;

  @encodedName("application/json", "star_task")
  starTask: string;

  @encodedName("application/json", "star_action")
  starAction: string;

  @encodedName("application/json", "star_result")
  starResult: string;

  keywords?: string[];
  source: string;

  @encodedName("application/json", "is_archived")
  isArchived: boolean;

  @encodedName("application/json", "created_at")
  createdAt: utcDateTime;

  @encodedName("application/json", "updated_at")
  updatedAt: utcDateTime;

  weapons?: ExperienceWeapon[];
}

model CreateExperienceRequest {
  title: string;
  category?: string;

  @encodedName("application/json", "period_start")
  periodStart?: string;

  @encodedName("application/json", "period_end")
  periodEnd?: string;

  role?: string;
  content?: string;
  result?: string;

  @encodedName("application/json", "star_situation")
  starSituation?: string;

  @encodedName("application/json", "star_task")
  starTask?: string;

  @encodedName("application/json", "star_action")
  starAction?: string;

  @encodedName("application/json", "star_result")
  starResult?: string;

  keywords?: string[];
}

model UpdateExperienceRequest {
  title?: string;
  category?: string;

  @encodedName("application/json", "period_start")
  periodStart?: string;

  @encodedName("application/json", "period_end")
  periodEnd?: string;

  role?: string;
  content?: string;
  result?: string;

  @encodedName("application/json", "star_situation")
  starSituation?: string;

  @encodedName("application/json", "star_task")
  starTask?: string;

  @encodedName("application/json", "star_action")
  starAction?: string;

  @encodedName("application/json", "star_result")
  starResult?: string;

  keywords?: string[];
}

model WeaponTag {
  code: string;
  confidence: float32;
  reasoning: string;
}

model WeaponTagResponse {
  @encodedName("application/json", "primary_weapon")
  primaryWeapon: WeaponTag;

  @encodedName("application/json", "secondary_weapons")
  secondaryWeapons: WeaponTag[];
}

// ─── Experience Interface ───

@route("/v1/experiences")
@tag("Experience")
@useAuth(BearerAuth)
interface ExperienceAPI {
  /** List experiences with filters */
  @get
  list(
    @query sort?: "latest" | "oldest" | "title",
    @query category?: string,
    @query weapon?: string,
  ): {
    @statusCode statusCode: 200;
    @body body: { experiences: Experience[] };
  } | Colight.Common.ErrorResponse;

  /** Create experience */
  @post
  create(@body body: CreateExperienceRequest): {
    @statusCode statusCode: 201;
    @body body: { id: string };
  } | Colight.Common.ErrorResponse;

  /** Get experience by ID */
  @get
  @route("/{id}")
  get(@path id: string): {
    @statusCode statusCode: 200;
    @body body: Experience;
  } | Colight.Common.ErrorResponse;

  /** Update experience */
  @patch(#{implicitOptionality: true})
  @route("/{id}")
  update(@path id: string, @body body: UpdateExperienceRequest): {
    @statusCode statusCode: 200;
    @body body: { id: string };
  } | Colight.Common.ErrorResponse;

  /** Delete experience */
  @delete
  @route("/{id}")
  delete(@path id: string): {
    @statusCode statusCode: 200;
    @body body: { success: boolean };
  } | Colight.Common.ErrorResponse;

  /** Auto-tag experience with weapons */
  @post
  @route("/{id}/tag")
  tag(@path id: string): {
    @statusCode statusCode: 200;
    @body body: WeaponTagResponse;
  } | Colight.Common.ErrorResponse;
}

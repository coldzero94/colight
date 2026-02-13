package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type PromptTemplate struct {
	ent.Schema
}

func (PromptTemplate) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
	}
}

func (PromptTemplate) Fields() []ent.Field {
	return []ent.Field{
		field.String("category").
			NotEmpty().
			MaxLen(50).
			Comment("Category: experience_classify, coaching_draft, etc."),
		field.String("sub_category").
			NotEmpty().
			MaxLen(50).
			Comment("Sub-category: weapon_tagging, interview, etc."),
		field.String("name").
			NotEmpty().
			MaxLen(100).
			Comment("Prompt name"),
		field.Text("system_prompt").
			NotEmpty().
			Comment("System prompt content"),
		field.Text("user_prompt_template").
			NotEmpty().
			Comment("User prompt with {{variable}} placeholders"),
		field.JSON("output_schema", map[string]interface{}{}).
			Optional().
			Comment("Expected output JSON schema"),
		field.String("model").
			NotEmpty().
			MaxLen(50).
			Comment("AI model: gemini-2.0-flash, groq-llama-3.3, claude-sonnet-4-5"),
		field.Float("temperature").
			Default(0.3).
			Comment("Model temperature"),
		field.Int("max_tokens").
			Default(2000).
			Comment("Max output tokens"),
		field.Int("version").
			Default(1).
			Comment("Prompt version number"),
		field.Bool("is_active").
			Default(true).
			Comment("Whether this prompt is active"),
		field.Int("usage_count").
			Default(0).
			Comment("Total usage count"),
		field.Int("avg_latency_ms").
			Default(0).
			Comment("Average latency in ms"),
		field.Float("avg_quality_score").
			Default(0).
			Comment("Average quality score"),
	}
}

func (PromptTemplate) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("coaching_sessions", CoachingSession.Type),
		edge.To("question_patterns", QuestionPattern.Type),
	}
}

func (PromptTemplate) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("category", "sub_category"),
		index.Fields("is_active", "category"),
	}
}

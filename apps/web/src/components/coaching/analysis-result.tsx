"use client";

import { IntentComparison } from "./intent-comparison";
import { WeaponBadges } from "./weapon-badges";
import { StructureChart } from "./structure-chart";
import { KeywordTags } from "./keyword-tags";
import { StructureExample } from "./structure-example";

interface RealIntent {
  intent: string;
  why: string;
}

interface WeaponInfo {
  weapon_id: string;
  weapon_name: string;
  reason: string;
}

interface Section {
  name: string;
  char_ratio: number;
  char_count: number;
  guide: string;
}

interface QuestionAnalysisResponse {
  surface_question: string;
  real_intents: RealIntent[];
  required_weapons: {
    primary: WeaponInfo;
    secondary: WeaponInfo[];
  };
  writing_structure: {
    total_chars: number;
    sections: Section[];
  };
  key_keywords: string[];
  avoid_list: string[];
  good_structure_example: string;
}

interface AnalysisResultProps {
  analysis: QuestionAnalysisResponse;
}

export function AnalysisResult({ analysis }: AnalysisResultProps) {
  return (
    <div className="space-y-6">
      {/* Intent comparison */}
      <IntentComparison
        surface={analysis.surface_question}
        intents={analysis.real_intents}
      />

      {/* Weapon badges */}
      <WeaponBadges
        primary={analysis.required_weapons.primary}
        secondary={analysis.required_weapons.secondary}
      />

      {/* Structure chart */}
      <StructureChart
        sections={analysis.writing_structure.sections}
        totalChars={analysis.writing_structure.total_chars}
      />

      {/* Keyword tags */}
      <KeywordTags
        keywords={analysis.key_keywords}
        avoidList={analysis.avoid_list}
      />

      {/* Structure example */}
      <StructureExample example={analysis.good_structure_example} />
    </div>
  );
}

"use client";

import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { cn } from "@/lib/utils";
import { Check } from "lucide-react";

export interface PricingCardProps {
  name: string;
  price: string;
  description: string;
  features: string[];
  cta: string;
  onCtaClick: () => void;
  highlighted?: boolean;
  current?: boolean;
}

export function PricingCard({
  name,
  price,
  description,
  features,
  cta,
  onCtaClick,
  highlighted,
  current,
}: PricingCardProps) {
  return (
    <Card
      className={cn(
        "flex flex-col",
        highlighted && "border-gray-900 shadow-lg",
      )}
    >
      <CardHeader>
        <CardTitle className="text-lg">{name}</CardTitle>
        <CardDescription>{description}</CardDescription>
        <p className="mt-2 text-3xl font-bold">{price}</p>
      </CardHeader>
      <CardContent className="flex-1">
        <ul className="space-y-2">
          {features.map((feature) => (
            <li key={feature} className="flex items-start gap-2 text-sm">
              <Check className="mt-0.5 h-4 w-4 shrink-0 text-green-600" />
              <span>{feature}</span>
            </li>
          ))}
        </ul>
      </CardContent>
      <CardFooter>
        <Button
          className="w-full"
          variant={highlighted ? "default" : "outline"}
          onClick={onCtaClick}
          disabled={current}
        >
          {current ? "현재 플랜" : cta}
        </Button>
      </CardFooter>
    </Card>
  );
}

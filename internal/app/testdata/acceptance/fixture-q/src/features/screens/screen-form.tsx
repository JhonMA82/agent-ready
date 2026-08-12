import { cn } from "@/lib/utils";

export function ScreenForm({ className }: { className?: string }) {
  return <form aria-label="screen form" className={cn("grid gap-4", className)} />;
}

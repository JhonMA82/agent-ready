import { cn } from "@/lib/utils";

export function ScreenList({ className }: { className?: string }) {
  return <ul aria-label="screen list" className={cn("divide-y", className)} />;
}

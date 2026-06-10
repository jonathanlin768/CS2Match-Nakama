import { Empty, EmptyHeader, EmptyTitle, EmptyDescription, EmptyMedia } from "@/components/ui/empty";
import { Ticket } from "lucide-react";

export default function CouponsTab() {
  return (
    <div className="space-y-6">
      <h3 className="text-lg font-semibold">我的代金券</h3>
      <Empty className="min-h-[320px] border-border border">
        <EmptyHeader>
          <EmptyMedia variant="icon">
            <Ticket className="w-6 h-6" />
          </EmptyMedia>
          <EmptyTitle>暂无代金券</EmptyTitle>
          <EmptyDescription>参与活动可获得代金券</EmptyDescription>
        </EmptyHeader>
      </Empty>
    </div>
  );
}

import { Empty, EmptyHeader, EmptyTitle, EmptyDescription, EmptyMedia } from "@/components/ui/empty";
import { ClipboardList } from "lucide-react";

export default function TasksTab() {
  return (
    <div className="space-y-6">
      <h3 className="text-lg font-semibold">我的任务</h3>
      <Empty className="min-h-[320px] border-border border">
        <EmptyHeader>
          <EmptyMedia variant="icon">
            <ClipboardList className="w-6 h-6" />
          </EmptyMedia>
          <EmptyTitle>暂无进行中的任务</EmptyTitle>
          <EmptyDescription>前往首页查看今日任务</EmptyDescription>
        </EmptyHeader>
      </Empty>
    </div>
  );
}

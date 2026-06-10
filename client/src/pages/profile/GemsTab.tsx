import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Empty, EmptyHeader, EmptyDescription } from "@/components/ui/empty";
import { Gem } from "lucide-react";
import { useOutletContext } from "react-router-dom";
import type { ProfileContext } from "./types";

export default function GemsTab() {
  const { userData } = useOutletContext<ProfileContext>();

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h3 className="text-lg font-semibold">我的钻石</h3>
        <Button size="sm" className="bg-primary hover:bg-primary/90">
          <Gem className="w-4 h-4 mr-1" />
          领取钻石
        </Button>
      </div>
      <Card className="bg-gradient-to-r from-pink-500/10 to-transparent border-pink-500/20">
        <CardContent className="py-6">
          <div className="flex items-center gap-3">
            <Gem className="w-8 h-8 text-pink-500" />
            <div>
              <p className="text-sm text-muted-foreground">当前余额</p>
              <p className="text-2xl font-bold">{userData.gems.toLocaleString()}</p>
            </div>
          </div>
        </CardContent>
      </Card>
      <Empty className="min-h-[240px] border-border border">
        <EmptyHeader>
          <EmptyDescription>暂无交易记录</EmptyDescription>
        </EmptyHeader>
      </Empty>
    </div>
  );
}

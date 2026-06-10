import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Crown } from "lucide-react";
import { useOutletContext } from "react-router-dom";
import type { ProfileContext } from "./types";

export default function VipTab() {
  const { userData } = useOutletContext<ProfileContext>();

  return (
    <div className="space-y-6">
      <h3 className="text-lg font-semibold">会员体系</h3>
      <Card className="bg-gradient-to-r from-primary/10 to-transparent border-primary/20">
        <CardContent className="py-8">
          <div className="flex items-center gap-4">
            <div className="w-16 h-16 rounded-full bg-primary/20 flex items-center justify-center">
              <Crown className="w-8 h-8 text-primary" />
            </div>
            <div>
              <div className="flex items-center gap-2">
                <span className="text-xl font-bold">VIP {userData.vipLevel}</span>
                <Badge variant="secondary">当前等级</Badge>
              </div>
              <p className="text-sm text-muted-foreground mt-1">
                再消费 5,000 钻石可升级至 VIP 4
              </p>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}

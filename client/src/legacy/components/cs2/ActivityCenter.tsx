"use client";

import { Gift, Zap, Star } from "lucide-react";

const activities = [
  {
    icon: Gift,
    title: "登录奖励",
    subtitle: "每日登录领取奖励",
    status: "已领取",
    statusColor: "text-muted-foreground",
  },
  {
    icon: Zap,
    title: "周末双倍",
    subtitle: "模拟对战奖励+100%",
    status: "进行中",
    statusColor: "text-green-400",
  },
  {
    icon: Star,
    title: "传奇降临",
    subtitle: "指定卡池概率UP",
    status: "2天后结束",
    statusColor: "text-primary",
  },
];

export function ActivityCenter() {
  return (
    <div className="flex-1 bg-card rounded-md border border-border p-3 lg:p-4 flex flex-col">
      <h3 className="font-semibold mb-3 lg:mb-4 text-sm lg:text-base">活动中心</h3>
      
      <div className="space-y-2 lg:space-y-3 flex-1">
        {activities.map((activity, index) => (
          <div
            key={index}
            className="flex items-center gap-2 lg:gap-3 p-1.5 lg:p-2 rounded hover:bg-secondary/50 transition-colors"
          >
            <div className="w-8 lg:w-10 h-8 lg:h-10 rounded bg-secondary flex items-center justify-center flex-shrink-0">
              <activity.icon className="w-4 lg:w-5 h-4 lg:h-5 text-primary" />
            </div>
            <div className="flex-1 min-w-0">
              <div className="text-xs lg:text-sm font-medium truncate">{activity.title}</div>
              <div className="text-[10px] lg:text-xs text-muted-foreground truncate">
                {activity.subtitle}
              </div>
            </div>
            <span className={`text-[10px] lg:text-xs flex-shrink-0 ${activity.statusColor}`}>
              {activity.status}
            </span>
          </div>
        ))}
      </div>

      <button className="w-full mt-3 lg:mt-4 py-1.5 lg:py-2 text-xs lg:text-sm text-muted-foreground hover:text-foreground transition-colors border border-border rounded">
        查看全部活动
      </button>
    </div>
  );
}

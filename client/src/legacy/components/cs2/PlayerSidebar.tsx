"use client";

import { Shield, Gift } from "lucide-react";

export function PlayerSidebar() {
  return (
    <div className="flex-1 flex flex-col sm:flex-row lg:flex-col gap-3">
      {/* Player Card */}
      <div className="bg-card rounded-md p-3 lg:p-4 border border-border flex-1 sm:flex-initial lg:flex-initial">
        <div className="flex items-center gap-3">
          <div className="relative">
            <div className="w-14 h-14 rounded-md bg-gradient-to-br from-blue-600 to-blue-800 flex items-center justify-center overflow-hidden">
              <div className="w-full h-full bg-[url('https://hebbkx1anhila5yf.public.blob.vercel-storage.com/file_00000000ab607209b3b7459ba45900da-H81WxYDQ5A5LWqbOB8WMNv6LIb1stU.png')] bg-cover bg-center opacity-80" />
            </div>
            <div className="absolute -bottom-1 -right-1 bg-blue-600 text-[10px] px-1.5 py-0.5 rounded text-white font-bold">
              S2
            </div>
          </div>
          <div>
            <div className="flex items-center gap-1">
              <span className="font-semibold">PlayerOne</span>
              <span className="text-primary text-xs">👑</span>
            </div>
            <div className="text-xs text-muted-foreground mt-1">
              12800 / 20000
            </div>
          </div>
        </div>
      </div>

      {/* Rank Card */}
      <div className="bg-card rounded-md p-3 lg:p-4 border border-border flex-1 sm:flex-initial lg:flex-initial">
        <div className="flex items-center gap-3">
          <div className="w-12 h-12 rounded-md bg-gradient-to-br from-amber-600 to-amber-800 flex items-center justify-center">
            <Shield className="w-7 h-7 text-amber-200" />
          </div>
          <div>
            <div className="font-semibold">大师守卫 II</div>
            <div className="text-xs text-muted-foreground mt-1">MMR 3621</div>
          </div>
        </div>
      </div>

      {/* Newbie Tasks - fills remaining height */}
      <div className="bg-card rounded-md border border-border flex-1 sm:flex-[2] lg:flex-1 flex flex-col overflow-hidden">
        {/* Task Image Background */}
        <div 
          className="h-24 bg-cover bg-center relative"
          style={{
            backgroundImage: "url('https://hebbkx1anhila5yf.public.blob.vercel-storage.com/file_00000000ab607209b3b7459ba45900da-H81WxYDQ5A5LWqbOB8WMNv6LIb1stU.png')",
          }}
        >
          <div className="absolute inset-0 bg-gradient-to-t from-card to-transparent" />
        </div>
        
        {/* Task Content */}
        <div className="p-3 flex-1 flex flex-col justify-between">
          <div className="flex items-center gap-2">
            <Gift className="w-5 h-5 text-primary" />
            <span className="font-semibold text-sm">新手任务</span>
          </div>
          <button className="w-full mt-2 text-xs text-primary border border-primary rounded-md py-1.5 hover:bg-primary/10 transition-colors">
            领取奖励
          </button>
        </div>
      </div>
    </div>
  );
}

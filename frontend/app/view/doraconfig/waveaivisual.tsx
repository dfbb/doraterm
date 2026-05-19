// Copyright 2025, Command Line Inc.
// SPDX-License-Identifier: Apache-2.0

import type { DoraConfigViewModel } from "@/app/view/doraconfig/doraconfig-model";
import { memo } from "react";

interface DoraAIVisualContentProps {
    model: DoraConfigViewModel;
}

export const DoraAIVisualContent = memo(({ model }: DoraAIVisualContentProps) => {
    return (
        <div className="flex flex-col gap-4 p-6 h-full">
            <div className="text-lg font-semibold">Dora AI Modes - Visual Editor</div>
            <div className="text-muted-foreground">Visual editor coming soon...</div>
        </div>
    );
});

DoraAIVisualContent.displayName = "DoraAIVisualContent";
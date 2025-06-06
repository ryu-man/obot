import { TriangleAlertIcon, WrenchIcon } from "lucide-react";

import { ToolReference } from "~/lib/model/toolReferences";

import { ToolIcon } from "~/components/tools/ToolIcon";
import { Button } from "~/components/ui/button";
import {
	Tooltip,
	TooltipContent,
	TooltipTrigger,
} from "~/components/ui/tooltip";

type ToolTooltipProps = {
	tool: ToolReference;
	children: React.ReactNode;
	onConfigureAuth?: () => void;
	requiresConfiguration?: boolean;
	isBundle?: boolean;
};

export function ToolTooltip({
	tool,
	children,
	onConfigureAuth,
	requiresConfiguration,
	isBundle = false,
}: ToolTooltipProps) {
	return (
		<Tooltip>
			<TooltipTrigger asChild>{children}</TooltipTrigger>
			<TooltipContent
				sideOffset={isBundle ? 255 : 30}
				side={isBundle ? "left" : "left"}
				className="flex w-[300px] items-center border bg-background p-4 text-foreground"
			>
				{tool.metadata?.icon ? (
					<ToolIcon
						icon={tool.metadata?.icon}
						category={tool.metadata?.category}
						name={tool.name}
						className="mr-4 h-10 w-10"
					/>
				) : (
					<WrenchIcon className="mr-2 h-4 w-4" />
				)}
				<div>
					<p className="font-bold">
						{tool.name}
						{isBundle ? " Bundle" : ""}
					</p>
					<p className="text-sm">
						{tool.description || "No description provided."}
					</p>
					{requiresConfiguration && onConfigureAuth && (
						<>
							<div className="flex items-center gap-1 pt-2 text-xs text-warning">
								<span>
									<TriangleAlertIcon className="h-4 w-4 text-warning" />
								</span>
								<p>
									<Button
										variant="link"
										className="p-0 text-xs"
										onClick={onConfigureAuth}
									>
										Setup
									</Button>{" "}
									required to use this tool.
								</p>
							</div>
						</>
					)}
				</div>
			</TooltipContent>
		</Tooltip>
	);
}

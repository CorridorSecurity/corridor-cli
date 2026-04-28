export interface MCPStatus {
  enabled: boolean;
  teamId?: string;
  teamName?: string;
  requiresCli: boolean;
}

export interface CheckUserMCPStatusResponse {
  mcpEnabled: boolean;
  teamId?: string;
  teamName?: string;
  requiresCli: boolean;
}

export async function checkUserMCPStatus(
  accessToken: string
): Promise<CheckUserMCPStatusResponse> {
  const response = await fetch("/mcp/me/mcp-enabled", {
    headers: {
      Authorization: `Bearer ${accessToken}`,
    },
  });

  if (!response.ok) {
    throw new Error(`Failed to check MCP status: ${response.statusText}`);
  }

  const data = await response.json();

  return {
    mcpEnabled: data.enabled ?? false,
    teamId: data.teamId,
    teamName: data.teamName,
    requiresCli: data.requiresCli ?? false,
  };
}

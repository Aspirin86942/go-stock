export function resolveFirstAiConfigId(aiConfigs) {
  if (!Array.isArray(aiConfigs) || aiConfigs.length === 0) {
    return null;
  }

  const firstConfig = aiConfigs[0];
  return firstConfig?.ID ?? firstConfig?.id ?? null;
}

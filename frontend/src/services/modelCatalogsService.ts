import { apiClient } from '../api/apiClient';

export type ModelDescriptor = {
  id: string;
  display_name: string;
  schema_version?: string;
  deprecated?: boolean;
};

export type ListModelCatalogResponse = {
  items: ModelDescriptor[];
  count: number;
};

export const modelCatalogsService = {
  async listAdvancementModels(): Promise<ListModelCatalogResponse> {
    return apiClient.get<ListModelCatalogResponse>('/advancement-models');
  },

  async listMarketShareModels(): Promise<ListModelCatalogResponse> {
    return apiClient.get<ListModelCatalogResponse>('/market-share-models');
  },

  async listEntryOptimizers(): Promise<ListModelCatalogResponse> {
    return apiClient.get<ListModelCatalogResponse>('/entry-optimizers');
  },
};

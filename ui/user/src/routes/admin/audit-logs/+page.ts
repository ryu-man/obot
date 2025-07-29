import { AdminService } from '$lib/services';
import type { PageLoad } from './$types';

export const load: PageLoad = ({ params, url }) => {
	const filters = url.searchParams.entries().reduce((acc, [key, value]) => {
		acc[key] = value;
		return acc;
	}, {});

	// const users = AdminService.listUsers();

	if (filters.mcpId) {
		return {
			// logs: AdminService.listServerOrInstanceAuditLogs(filters.mcpId, filters)
			// users
		};
	}

	return {
		// logs: AdminService.listAuditLogs(filters),
		// users
	};
};

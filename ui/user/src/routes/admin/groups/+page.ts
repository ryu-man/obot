import { handleRouteError } from '$lib/errors';
import { AdminService, type OrgGroup } from '$lib/services';
import { profile } from '$lib/stores';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ fetch }) => {
	const allGroupsPromise = new Promise<OrgGroup[]>(async (resolve, reject) => {
		try {
			const groups = await AdminService.listGroups({ fetch, includeRestricted: true });

			resolve([...groups]);
		} catch (err) {
			reject(err);
		}
	});

	try {
		return {
			groups: await allGroupsPromise,
			groupRoleAssignments: await AdminService.listGroupRoleAssignments({ fetch })
		};
	} catch (err) {
		handleRouteError(err, `/admin/groups`, profile.current);

		return {
			groups: [],
			groupRoleAssignments: []
		};
	}
};

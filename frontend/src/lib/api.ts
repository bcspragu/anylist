const postData = (path: string, data: FormData): Promise<Response> => {
	return fetch(path, {
		method: 'POST',
		body: data
	});
};

export const addItem = (itemName: string): Promise<Response> => {
	const formData = new FormData();
	formData.append('item_name', itemName);
	return postData('/api/add', formData);
};

export const removeItem = (itemID: string): Promise<Response> => {
	const formData = new FormData();
	formData.append('item_id', itemID);
	return postData('/api/remove', formData);
};

export const checkItem = (itemID: string, checked: boolean): Promise<Response> => {
	const formData = new FormData();
	formData.append('item_id', itemID);
	formData.append('checked', checked ? 'true' : 'false');
	return postData('/api/check', formData);
};

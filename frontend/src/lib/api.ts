const postData = (url: string, data: FormData): Promise<Response> => {
	return fetch(url, {
		method: 'POST',
		body: data
	});
};

export const addItem = (itemName: string): Promise<Response> => {
	const formData = new FormData();
	formData.append('item_name', itemName);
	return postData('http://localhost:8080/api/add', formData);
};

export const removeItem = (itemID: string): Promise<Response> => {
	const formData = new FormData();
	formData.append('item_id', itemID);
	return postData('http://localhost:8080/api/remove', formData);
};

export const checkItem = (itemID: string, checked: boolean): Promise<Response> => {
	const formData = new FormData();
	formData.append('item_id', itemID);
	formData.append('checked', checked ? 'true' : 'false');
	return postData('http://localhost:8080/api/check', formData);
};

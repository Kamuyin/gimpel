import BaseComponent from "sap/ui/core/UIComponent";
import { createDeviceModel } from "./model/models";
import APIService from "./service/APIService";

/**
 * @namespace io.github.kamuyin.gimpel.web
 */
export default class Component extends BaseComponent {

	public static metadata = {
		manifest: "json",
        interfaces: [
            "sap.ui.core.IAsyncContentCreation"
        ]
	};

	private apiService: APIService;

	public init() : void {
		// call the base component's init function
		super.init();

        // set the device model
        this.setModel(createDeviceModel(), "device");

		this.apiService = new APIService("/api/v1");

        // enable routing
        this.getRouter().initialize();
	}

	public getAPIService(): APIService {
		return this.apiService;
	}
}
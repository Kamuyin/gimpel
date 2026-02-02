import Component from "../Component";
import APIService from "../service/APIService";
import Controller from "sap/ui/core/mvc/Controller";


export function getAPIService(controller: Controller): APIService {
    const component = controller.getOwnerComponent() as Component;
    return component.getAPIService();
}

// *** WARNING: this file was generated by the Pulumi SDK Generator. ***
// *** Do not edit by hand unless you're certain you know what you are doing! ***

import * as pulumi from "@pulumi/pulumi";
import * as utilities from "./utilities";

export class Project extends pulumi.CustomResource {
    /**
     * Get an existing Project resource's state with the given name, ID, and optional extra
     * properties used to qualify the lookup.
     *
     * @param name The _unique_ name of the resulting resource.
     * @param id The _unique_ provider ID of the resource to lookup.
     * @param opts Optional settings to control the behavior of the CustomResource.
     */
    public static get(name: string, id: pulumi.Input<pulumi.ID>, opts?: pulumi.CustomResourceOptions): Project {
        return new Project(name, undefined as any, { ...opts, id: id });
    }

    /** @internal */
    public static readonly __pulumiType = 'sentry:index:Project';

    /**
     * Returns true if the given object is an instance of Project.  This is designed to work even
     * when multiple copies of the Pulumi SDK have been loaded into the same process.
     */
    public static isInstance(obj: any): obj is Project {
        if (obj === undefined || obj === null) {
            return false;
        }
        return obj['__pulumiType'] === Project.__pulumiType;
    }

    public readonly name!: pulumi.Output<string>;
    public readonly organizationSlug!: pulumi.Output<string>;
    public readonly slug!: pulumi.Output<string | undefined>;
    public readonly teamSlug!: pulumi.Output<string>;

    /**
     * Create a Project resource with the given unique name, arguments, and options.
     *
     * @param name The _unique_ name of the resource.
     * @param args The arguments to use to populate this resource's properties.
     * @param opts A bag of options that control this resource's behavior.
     */
    constructor(name: string, args: ProjectArgs, opts?: pulumi.CustomResourceOptions) {
        let inputs: pulumi.Inputs = {};
        if (!(opts && opts.id)) {
            if (!args || args.name === undefined) {
                throw new Error("Missing required property 'name'");
            }
            if (!args || args.organizationSlug === undefined) {
                throw new Error("Missing required property 'organizationSlug'");
            }
            if (!args || args.slug === undefined) {
                throw new Error("Missing required property 'slug'");
            }
            if (!args || args.teamSlug === undefined) {
                throw new Error("Missing required property 'teamSlug'");
            }
            inputs["name"] = args ? args.name : undefined;
            inputs["organizationSlug"] = args ? args.organizationSlug : undefined;
            inputs["slug"] = args ? args.slug : undefined;
            inputs["teamSlug"] = args ? args.teamSlug : undefined;
        } else {
            inputs["name"] = undefined /*out*/;
            inputs["organizationSlug"] = undefined /*out*/;
            inputs["slug"] = undefined /*out*/;
            inputs["teamSlug"] = undefined /*out*/;
        }
        if (!opts) {
            opts = {}
        }

        if (!opts.version) {
            opts.version = utilities.getVersion();
        }
        super(Project.__pulumiType, name, inputs, opts);
    }
}

/**
 * The set of arguments for constructing a Project resource.
 */
export interface ProjectArgs {
    readonly name: pulumi.Input<string>;
    readonly organizationSlug: pulumi.Input<string>;
    readonly slug: pulumi.Input<string>;
    readonly teamSlug: pulumi.Input<string>;
}

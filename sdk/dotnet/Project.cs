// *** WARNING: this file was generated by the Pulumi SDK Generator. ***
// *** Do not edit by hand unless you're certain you know what you are doing! ***

using System;
using System.Collections.Generic;
using System.Collections.Immutable;
using System.Threading.Tasks;
using Pulumi.Serialization;

namespace Pulumi.Sentry
{
    public partial class Project : Pulumi.CustomResource
    {
        [Output("defaultClientKeyDSNPublic")]
        public Output<string?> DefaultClientKeyDSNPublic { get; private set; } = null!;

        [Output("defaultEnvironment")]
        public Output<string?> DefaultEnvironment { get; private set; } = null!;

        [Output("name")]
        public Output<string> Name { get; private set; } = null!;

        [Output("organizationSlug")]
        public Output<string?> OrganizationSlug { get; private set; } = null!;

        [Output("slug")]
        public Output<string> Slug { get; private set; } = null!;

        [Output("subjectPrefix")]
        public Output<string?> SubjectPrefix { get; private set; } = null!;

        [Output("subjectTemplate")]
        public Output<string?> SubjectTemplate { get; private set; } = null!;

        [Output("teamSlug")]
        public Output<string?> TeamSlug { get; private set; } = null!;


        /// <summary>
        /// Create a Project resource with the given unique name, arguments, and options.
        /// </summary>
        ///
        /// <param name="name">The unique name of the resource</param>
        /// <param name="args">The arguments used to populate this resource's properties</param>
        /// <param name="options">A bag of options that control this resource's behavior</param>
        public Project(string name, ProjectArgs args, CustomResourceOptions? options = null)
            : base("sentry:index:Project", name, args ?? new ProjectArgs(), MakeResourceOptions(options, ""))
        {
        }

        private Project(string name, Input<string> id, CustomResourceOptions? options = null)
            : base("sentry:index:Project", name, null, MakeResourceOptions(options, id))
        {
        }

        private static CustomResourceOptions MakeResourceOptions(CustomResourceOptions? options, Input<string>? id)
        {
            var defaultOptions = new CustomResourceOptions
            {
                Version = Utilities.Version,
            };
            var merged = CustomResourceOptions.Merge(defaultOptions, options);
            // Override the ID if one was specified for consistency with other language SDKs.
            merged.Id = id ?? merged.Id;
            return merged;
        }
        /// <summary>
        /// Get an existing Project resource's state with the given name, ID, and optional extra
        /// properties used to qualify the lookup.
        /// </summary>
        ///
        /// <param name="name">The unique name of the resulting resource.</param>
        /// <param name="id">The unique provider ID of the resource to lookup.</param>
        /// <param name="options">A bag of options that control this resource's behavior</param>
        public static Project Get(string name, Input<string> id, CustomResourceOptions? options = null)
        {
            return new Project(name, id, options);
        }
    }

    public sealed class ProjectArgs : Pulumi.ResourceArgs
    {
        [Input("defaultEnvironment")]
        public Input<string>? DefaultEnvironment { get; set; }

        [Input("name", required: true)]
        public Input<string> Name { get; set; } = null!;

        [Input("organizationSlug", required: true)]
        public Input<string> OrganizationSlug { get; set; } = null!;

        [Input("slug", required: true)]
        public Input<string> Slug { get; set; } = null!;

        [Input("subjectPrefix")]
        public Input<string>? SubjectPrefix { get; set; }

        [Input("subjectTemplate")]
        public Input<string>? SubjectTemplate { get; set; }

        [Input("teamSlug", required: true)]
        public Input<string> TeamSlug { get; set; } = null!;

        public ProjectArgs()
        {
        }
    }
}
